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
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"heimdall/app/internal/hub"
	"heimdall/app/internal/secure"
	v1 "heimdall/common/proto/monitoring/v1"
)

func main() {
	listen := flag.String("listen", ":9090", "gRPC listen address")
	staleAfter := flag.Duration("stale-after", 10*time.Second, "mark a host stale after no updates for this long")
	offlineAfter := flag.Duration("offline-after", 30*time.Second, "mark a host offline after no updates for this long")
	token := flag.String("token", os.Getenv("HEIMDALL_TOKEN"), "required enrollment token (env HEIMDALL_TOKEN); empty disables auth")
	tlsCert := flag.String("tls-cert", "", "PEM server certificate; enables TLS with --tls-key")
	tlsKey := flag.String("tls-key", "", "PEM server private key; enables TLS with --tls-cert")
	id := flag.String("id", defaultHubID(), "this hub's federation id (origin of local hosts, appended to relay paths)")
	upstream := flag.String("upstream", "", "parent hub address to relay this hub's hosts to (federation)")
	upstreamToken := flag.String("upstream-token", os.Getenv("HEIMDALL_UPSTREAM_TOKEN"), "enrollment token for the upstream hub (env HEIMDALL_UPSTREAM_TOKEN)")
	upstreamTLS := flag.Bool("upstream-tls", false, "relay to the upstream hub over TLS")
	upstreamCA := flag.String("upstream-tls-ca", "", "PEM CA bundle to trust for the upstream hub")
	upstreamServerName := flag.String("upstream-tls-server-name", "", "override the server name verified in the upstream certificate")
	upstreamInsecure := flag.Bool("upstream-tls-insecure", false, "skip upstream certificate verification (dev only)")
	relayInterval := flag.Duration("relay-interval", 2*time.Second, "how often to relay hosts upstream")
	flag.Parse()

	h := hub.New(*staleAfter, *offlineAfter)
	h.SetToken(*token)
	h.SetID(*id)

	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		fail(err)
	}

	creds, err := secure.ServerOption(*tlsCert, *tlsKey)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.EvaluateLoop(ctx)

	if *upstream != "" {
		dialOpts, err := upstreamDialOptions(*upstreamToken, secure.ClientConfig{
			Enabled: *upstreamTLS, CAFile: *upstreamCA, ServerName: *upstreamServerName, SkipVerify: *upstreamInsecure,
		})
		if err != nil {
			fail(err)
		}
		go hub.RunRelay(ctx, h, *upstream, dialOpts, *relayInterval)
		fmt.Fprintf(os.Stderr, "heimdall-hub: relaying upstream to %s every %s\n", *upstream, *relayInterval)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Fprintln(os.Stderr, "heimdall-hub: shutting down")
		srv.GracefulStop()
	}()

	fmt.Fprintf(os.Stderr, "heimdall-hub: id=%s listening on %s (stale %s, offline %s, tls=%t, auth=%t)\n",
		*id, *listen, *staleAfter, *offlineAfter, *tlsCert != "", *token != "")
	if err := srv.Serve(lis); err != nil {
		fail(err)
	}
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
