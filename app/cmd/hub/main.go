// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Command heimdall-hub is the central gRPC server: daemons stream metrics to it
// and dashboards subscribe for live fan-out.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"heimdall/app/internal/hub"
	"heimdall/app/internal/observe"
	"heimdall/app/internal/options"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/selfupdate"
	v1 "heimdall/common/proto/monitoring/v1"
)

// version is the Heimdall build version, set via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := selfupdate.Run("hub", version); err != nil {
			fmt.Fprintln(os.Stderr, "heimdall-hub:", err)
			os.Exit(1)
		}
		return
	}
	showVersion := flag.Bool("version", false, "print version and exit")
	cat := hubCatalog()
	cat.Register(flag.CommandLine)
	flag.Usage = usage
	flag.Parse()
	if *showVersion {
		fmt.Println("heimdall-hub", version)
		return
	}

	cfg := options.Resolve("hub", cat,
		"heimdall-hub — first-run setup",
		"The hub ingests metrics from daemons and fans them out to dashboards.")

	h := hub.New(cfg.Span("stale-after", 10*time.Second), cfg.Span("offline-after", 30*time.Second))
	h.SetToken(cfg.Secret("token").Reveal())
	h.SetID(cfg.Text("id"))
	h.Registry().SetPurgeAfter(cfg.Span("purge-after", 15*time.Minute))
	h.Registry().SetHubLabels(options.ParseTags(cfg.Text("tags")))

	creds, err := secure.ServerOption(cfg.Text("tls-cert"), cfg.Text("tls-key"))
	if err != nil {
		fail(err)
	}
	srv := grpc.NewServer(
		creds,
		grpc.UnaryInterceptor(h.UnaryInterceptor()),
		grpc.StreamInterceptor(h.StreamInterceptor()),
	)
	v1.RegisterEnrollmentServiceServer(srv, h)
	v1.RegisterMetricStreamServiceServer(srv, h)
	v1.RegisterFederationServiceServer(srv, h)

	addrs := listenAddrs(cfg.Text("listen"))
	listeners := make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			fail(err)
		}
		listeners = append(listeners, lis)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.EvaluateLoop(ctx)

	if upstream := cfg.Text("upstream"); upstream != "" {
		dialOpts, err := upstreamDialOptions(cfg.Secret("upstream-token").Reveal(), secure.ClientConfig{
			Enabled: cfg.Toggle("upstream-tls"), CAFile: cfg.Text("upstream-tls-ca"),
			ServerName: cfg.Text("upstream-tls-server-name"), SkipVerify: cfg.Toggle("upstream-tls-insecure"),
		})
		if err != nil {
			fail(err)
		}
		interval := cfg.Span("relay-interval", 2*time.Second)
		go hub.RunRelay(ctx, h, upstream, dialOpts, interval)
		fmt.Fprintf(os.Stderr, "heimdall-hub: relaying upstream to %s every %s\n", upstream, interval)
	}

	// Mímir: serve Prometheus/OpenMetrics and keep bounded in-memory history.
	if mAddr := cfg.Text("metrics-listen"); mAddr != "" {
		hist := observe.NewHistory(1000)
		go func() {
			t := time.NewTicker(5 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case now := <-t.C:
					hist.Record(h.Registry().Hosts(), now)
				}
			}
		}()
		ms := &http.Server{Addr: mAddr, Handler: observe.Handler(h.Registry(), hist)}
		go func() {
			if err := ms.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintln(os.Stderr, "heimdall-hub: metrics server:", err)
			}
		}()
		go func() { <-ctx.Done(); _ = ms.Close() }()
		fmt.Fprintf(os.Stderr, "heimdall-hub: serving Prometheus metrics on %s (Mímir)\n", mAddr)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Fprintln(os.Stderr, "heimdall-hub: shutting down")
		// The metric-stream and subscribe RPCs are long-lived and never return on
		// their own, so GracefulStop blocks for as long as any daemon or dashboard
		// stays connected. Give in-flight work a brief grace period, then force
		// every stream closed so the hub always exits — connected or not.
		stopped := make(chan struct{})
		go func() { srv.GracefulStop(); close(stopped) }()
		select {
		case <-stopped:
		case <-time.After(2 * time.Second):
			srv.Stop()
		}
	}()

	fmt.Fprintf(os.Stderr, "heimdall-hub: id=%s listening on %s (tls=%t, auth=%t)\n",
		cfg.Text("id"), strings.Join(addrs, ","), cfg.Text("tls-cert") != "", !cfg.Secret("token").IsEmpty())

	var wg sync.WaitGroup
	for _, lis := range listeners {
		wg.Add(1)
		go func(l net.Listener) {
			defer wg.Done()
			if err := srv.Serve(l); err != nil {
				fmt.Fprintln(os.Stderr, "heimdall-hub:", err)
			}
		}(lis)
	}
	wg.Wait()
}

// defaultHubID derives a stable-ish federation id from the hostname.
func defaultHubID() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "hub"
}

// upstreamDialOptions assembles transport security and the per-RPC token for the
// cross-hub relay link.
func upstreamDialOptions(token string, tlsCfg secure.ClientConfig) ([]grpc.DialOption, error) {
	transportOpt, err := tlsCfg.DialOption()
	if err != nil {
		return nil, err
	}
	opts := []grpc.DialOption{transportOpt}
	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(secure.TokenCredentials{Token: token}))
	}
	return opts, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "heimdall-hub:", err)
	os.Exit(1)
}
