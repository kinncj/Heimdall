// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Command heimdall-daemon samples this host's metrics through the unprivileged
// adapters. With --hub it streams them to a hub over the versioned gRPC contract
// (auto-reconnecting); without it, it prints samples (text or --json).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"heimdall/app/internal/adapters"
	"heimdall/app/internal/command"
	"heimdall/app/internal/discovery"
	"heimdall/app/internal/domain"
	"heimdall/app/internal/helper"
	"heimdall/app/internal/logs"
	"heimdall/app/internal/options"
	"heimdall/app/internal/proc"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/selfupdate"
	"heimdall/app/internal/transport"
	v1 "heimdall/common/proto/monitoring/v1"
)

// logger carries the daemon's structured operational logs. main replaces it
// once --log-file is parsed; the default keeps early errors on stderr.
// version is the Heimdall build version, set via -ldflags "-X main.version=…"
// at release time; "dev" for local builds. Reported by the Inventory adapter.
var version = "dev"

var logger = slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("component", "heimdall-daemon")

func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := selfupdate.Run("daemon", version); err != nil {
			fmt.Fprintln(os.Stderr, "heimdall-daemon:", err)
			os.Exit(1)
		}
		return
	}
	showVersion := flag.Bool("version", false, "print version and exit")
	once := flag.Bool("once", false, "collect a single sample and exit (print mode; not saved)")
	printLocal := flag.Bool("print", false, "print samples locally instead of streaming (not saved)")
	cat := daemonCatalog()
	cat.Register(flag.CommandLine)
	flag.Usage = usage
	flag.Parse()
	if *showVersion {
		fmt.Println("heimdall-daemon", version)
		return
	}

	cfg := resolveDaemon(cat)

	lg, metricsW, metricsJSON, closeLog, err := resolveOutput(cfg.Text("log-file"), cfg.Toggle("json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "heimdall-daemon:", err)
		os.Exit(1)
	}
	defer closeLog()
	logger = lg

	host := cfg.Text("name")
	if host == "" {
		if h, err := os.Hostname(); err == nil && h != "" {
			host = h
		} else {
			host = "localhost"
		}
	}

	interval := cfg.Span("interval", 2*time.Second)
	tags := options.ParseTags(cfg.Text("tags"))
	reg := domain.NewRegistry(interval)
	for _, a := range adapters.Build(adapters.Options{PingTarget: cfg.Text("ping-target"), Version: version}) {
		reg.Register(a)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	hubAddr := cfg.Address("hub")
	hubTarget := hubAddr.String()
	// Ratatoskr: discover the hub when --hub is "auto", or when --discover is set
	// and no explicit --hub was given. An explicit --hub always wins.
	discover := cfg.Text("hub") == "auto" || (cfg.Toggle("discover") && hubAddr.IsEmpty())
	streaming := !*printLocal && !*once && (discover || !hubAddr.IsEmpty())
	if streaming && discover {
		addr, err := discoverHub(cfg.Text("discover-seed"))
		if err != nil {
			fatal("hub discovery failed (Ratatoskr)", err)
		}
		hubTarget = addr
		logger.Info("discovered hub", "addr", addr)
	}
	if streaming {
		dialOpts, err := clientDialOptions(cfg.Secret("token").Reveal(), secure.ClientConfig{
			Enabled: cfg.Toggle("tls"), CAFile: cfg.Text("tls-ca"),
			ServerName: cfg.Text("tls-server-name"), SkipVerify: cfg.Toggle("tls-insecure"),
		})
		if err != nil {
			fatal("invalid client TLS configuration", err)
		}
		// Heimdallr's sight (ADR 0017): tail log sources and collect a periodic
		// process table, pushed to the hub on the existing stream. The reserved
		// labels advertise the capability. The daemon never listens.
		sources, err := logs.ParseSources(cfg.Text("log-source"))
		if err != nil {
			fatal("invalid --log-source", err)
		}
		push, pushLabels := startPush(context.Background(), sources, cfg.Span("process-interval", 0), proc.Local{}, cfg.Toggle("allow-commands"))
		if len(pushLabels) > 0 && tags == nil {
			tags = map[string]string{} // ParseTags returns nil when no --tags were given
		}
		for k, v := range pushLabels {
			tags[k] = v
		}
		streamToHub(hubTarget, host, interval, tags, reg, sig, dialOpts, push)
		return
	}

	sample := func() { emit(metricsW, host, reg.Collect(context.Background()), metricsJSON) }
	if *once {
		sample()
		return
	}
	logger.Info("daemon started",
		"host", host, "os", runtime.GOOS, "arch", runtime.GOARCH, "interval", interval.String())
	t := time.NewTicker(interval)
	defer t.Stop()
	sample()
	for {
		select {
		case <-t.C:
			sample()
		case <-sig:
			logger.Info("shutting down")
			return
		}
	}
}

// resolveOutput maps --log-file to where the daemon sends its metric samples and
// its structured logs:
//
//   - unset (default): metric samples to stdout, JSON logs to stderr (the TTY)
//   - "false"/off/none/0: both discarded — no output whatsoever
//   - a path: both written to that file as JSON lines (aggregator-friendly)
//
// In file mode the metric samples and logs share one synchronized writer so
// concurrent lines never interleave, and metric output is forced to JSON.
func resolveOutput(logFile string, jsonFlag bool) (lg *slog.Logger, metricsW io.Writer, metricsJSON bool, closeFn func() error, err error) {
	noop := func() error { return nil }
	switch strings.ToLower(strings.TrimSpace(logFile)) {
	case "":
		return newJSONLogger(os.Stderr), os.Stdout, jsonFlag, noop, nil
	case "false", "off", "none", "disabled", "0":
		return newJSONLogger(io.Discard), io.Discard, false, noop, nil
	default:
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, false, nil, fmt.Errorf("open log file %q: %w", logFile, err)
		}
		w := &syncWriter{w: f}
		return newJSONLogger(w), w, true, f.Close, nil
	}
}

func newJSONLogger(w io.Writer) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(h).With("component", "heimdall-daemon")
}

// syncWriter serializes writes so that metric samples and log records sharing a
// single log file never interleave mid-line.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

// fatal logs a structured error and exits non-zero.
func fatal(msg string, err error) {
	logger.Error(msg, "error", err.Error())
	os.Exit(1)
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, `heimdall-daemon — Heimdall metrics daemon

Samples this host's metrics (CPU, per-core, memory, disk, temperature, network
throughput, internet + per-NIC gateway latency, uptime, and GPU/power where
available). With --hub it streams them to a hub over gRPC (auto-reconnecting);
without it, it prints samples.

Usage:
  heimdall-daemon [flags]

Examples:
  heimdall-daemon                                 print this host's metrics
  heimdall-daemon --once --json                   one sample as JSON, then exit
  heimdall-daemon --hub localhost:9090 --name web stream to a hub
  heimdall-daemon --ping-target 8.8.8.8           use a different reachability target
  heimdall-daemon --hub h:9090 --log-source app=/var/log/app.log   push logs to the hub
  heimdall-daemon --hub h:9090 --process-interval 5s               push a process table for top
  heimdall-daemon --log-file /var/log/heimdall/daemon.json

Flags:
`)
	flag.PrintDefaults()
}

// clientDialOptions assembles the transport security and per-RPC enrollment
// token used for every hub connection.
func clientDialOptions(token string, tlsCfg secure.ClientConfig) ([]grpc.DialOption, error) {
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

// runCommand executes an allow-listed, read-only on-demand command (v2 Phase 2,
// ADR 0018), audits it, stages the result for the next snapshot, and triggers an
// immediate send. Commands run as this unprivileged daemon user.
func runCommand(req *v1.ControlRequest, push *pusher, trigger chan struct{}) {
	if !push.allowCommands {
		logger.Warn("control command refused: commands disabled", "cmd", req.GetAllowlistedCmd(), "actor", req.GetActor())
		push.setResult(&v1.ControlResponse{
			RequestId: req.GetRequestId(),
			ExitCode:  -1,
			Stderr:    "on-demand commands are disabled on this host (start the daemon with --allow-commands)",
			Status:    v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION,
		})
		select {
		case trigger <- struct{}{}:
		default:
		}
		return
	}
	key := req.GetAllowlistedCmd()
	var res command.Result
	if command.IsPrivileged(key) {
		// Privileged commands are delegated to the root helper over the local unix
		// socket (v2 Phase 2b). The unprivileged daemon never gains privilege; a
		// host with no helper simply cannot satisfy the request.
		r, err := (helper.Client{}).Exec(context.Background(), key, req.GetArgs())
		switch {
		case errors.Is(err, helper.ErrUnavailable):
			res = command.Result{Status: domain.StatusInsufficientPermission, ExitCode: -1,
				Stderr: "this command needs the privileged helper (heimdall-helper), which is not running on this host"}
		case err != nil:
			res = command.Result{Status: domain.StatusError, ExitCode: -1, Stderr: err.Error()}
		default:
			res = r
		}
	} else {
		res = command.Run(context.Background(), key, req.GetArgs())
	}
	logger.Info("control command",
		"cmd", key, "args", req.GetArgs(), "privileged", command.IsPrivileged(key),
		"actor", req.GetActor(), "status", res.Status.String(), "exit", res.ExitCode)
	push.setResult(&v1.ControlResponse{
		RequestId: req.GetRequestId(),
		ExitCode:  int32(res.ExitCode),
		Stdout:    res.Stdout,
		Stderr:    res.Stderr,
		Truncated: res.Truncated,
		Status:    transport.StatusToProto(res.Status),
	})
	select {
	case trigger <- struct{}{}:
	default:
	}
}

// discoverHub resolves a hub address via mDNS (Ratatoskr), falling back to a
// static seed for overlay networks where multicast does not reach.
func discoverHub(seed string) (string, error) {
	return discovery.Resolve(seed, 5*time.Second)
}

// streamToHub streams snapshots to the hub, reconnecting with backoff on error.
func streamToHub(addr, host string, interval time.Duration, labels map[string]string, reg *domain.Registry, sig chan os.Signal, dialOpts []grpc.DialOption, push *pusher) {
	backoff := time.Second
	for {
		connected, err := streamOnce(addr, host, interval, labels, reg, sig, dialOpts, push)
		if err == nil {
			return // clean shutdown
		}
		if connected {
			backoff = time.Second // reset after a healthy session
		}
		logger.Warn("hub connection lost; reconnecting",
			"error", err.Error(), "backoff", backoff.String())
		select {
		case <-time.After(backoff):
		case <-sig:
			return
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func streamOnce(addr, host string, interval time.Duration, labels map[string]string, reg *domain.Registry, sig chan os.Signal, dialOpts []grpc.DialOption, push *pusher) (bool, error) {
	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return false, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	ctx := context.Background()
	enrollReq := &v1.EnrollRequest{Host: &v1.Host{
		HostId: host, Hostname: host, DisplayName: host,
		Context: &v1.HostContext{Os: runtime.GOOS, Arch: runtime.GOARCH, Labels: labels},
	}}
	if _, err := v1.NewEnrollmentServiceClient(conn).Enroll(ctx, enrollReq); err != nil {
		return false, fmt.Errorf("enroll: %w", err)
	}
	stream, err := v1.NewMetricStreamServiceClient(conn).Stream(ctx)
	if err != nil {
		return false, fmt.Errorf("open stream: %w", err)
	}
	logger.Info("streaming to hub", "addr", addr, "host", host)

	// Heimdallr's sight v2 (ADR 0018): receive directives down the same outbound
	// stream — demand windows and on-demand allow-listed commands — and act on
	// them. The daemon still never listens; the hub drives it over the daemon's own
	// connection. A finished command triggers an immediate send so the result is
	// timely rather than waiting for the next tick.
	trigger := make(chan struct{}, 1)
	if push != nil {
		go func() {
			for {
				ctrl, err := stream.Recv()
				if err != nil {
					return // stream closed; streamOnce will reconnect
				}
				if w := ctrl.GetObservability(); w != nil {
					push.setWindow(w.GetLogs(), w.GetProcesses())
				}
				if req := ctrl.GetRun(); req != nil {
					go runCommand(req, push, trigger)
				}
			}
		}()
	}

	var seq uint64
	send := func() error {
		seq++
		snap := transport.ToSnapshot(host, reg.Collect(ctx), labels, seq, time.Now())
		if push != nil {
			procs, procsAt, lines := push.drain()
			transport.AttachObservability(snap, procs, procsAt, host, lines)
			snap.CommandResult = push.drainResult()
		}
		return stream.Send(snap)
	}
	if err := send(); err != nil {
		return true, fmt.Errorf("send: %w", err)
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := send(); err != nil {
				return true, fmt.Errorf("send: %w", err)
			}
		case <-trigger:
			// Flush every queued command result — one snapshot each, since the wire
			// carries a single result per snapshot — so concurrent commands don't
			// wait a tick apiece (or, before the queue, overwrite each other).
			for {
				if err := send(); err != nil {
					return true, fmt.Errorf("send: %w", err)
				}
				if push == nil || !push.hasResults() {
					break
				}
			}
		case <-sig:
			_ = stream.CloseSend()
			logger.Info("shutting down")
			return true, nil
		}
	}
}

func emit(w io.Writer, host string, ms []domain.Metric, asJSON bool) {
	if asJSON {
		type row struct {
			Host   string    `json:"host"`
			TS     int64     `json:"ts"`
			Metric string    `json:"metric"`
			Status string    `json:"status"`
			Value  float64   `json:"value"`
			Unit   string    `json:"unit"`
			Detail string    `json:"detail,omitempty"`
			Cores  []float64 `json:"cores,omitempty"`
		}
		enc := json.NewEncoder(w)
		now := time.Now().Unix()
		for _, m := range ms {
			_ = enc.Encode(row{host, now, m.Name, m.Status.String(), m.Gauge, m.Unit, m.Detail, m.PerCore})
		}
		return
	}
	line := host
	for _, m := range ms {
		line += "  " + formatMetric(m)
	}
	fmt.Fprintln(w, line)
}

// unitValue renders the value part (no name) of an OK gauge metric, keyed by
// unit. A unit with no entry falls back to defaultValue. Per-core and non-OK
// metrics never reach this table — their value does not come from Gauge.
var unitValue = map[string]func(domain.Metric) string{
	"info":    func(m domain.Metric) string { return m.Detail },
	"percent": func(m domain.Metric) string { return fmt.Sprintf("%.0f%%", m.Gauge) },
	"watts":   func(m domain.Metric) string { return fmt.Sprintf("%.2fW", m.Gauge) },
	"celsius": func(m domain.Metric) string { return fmt.Sprintf("%.0f°C", m.Gauge) },
	"ms":      func(m domain.Metric) string { return fmt.Sprintf("%.0fms", m.Gauge) },
	"MB/s":    func(m domain.Metric) string { return fmt.Sprintf("%.2fMB/s", m.Gauge) },
	"s":       func(m domain.Metric) string { return shortDuration(m.Gauge) },
}

func defaultValue(m domain.Metric) string { return fmt.Sprintf("%g", m.Gauge) }

// formatMetric renders one metric for the human-readable print line as
// "name=value". Non-OK metrics show their status; per-core metrics show an
// avg/max summary; everything else delegates the value to a per-unit formatter.
func formatMetric(m domain.Metric) string {
	if m.Status != domain.StatusOK {
		return fmt.Sprintf("%s=%s", m.Name, m.Status)
	}
	if m.Kind == domain.KindPerCore && len(m.PerCore) > 0 {
		return fmt.Sprintf("%s=%s", m.Name, perCoreValue(m))
	}
	value, ok := unitValue[m.Unit]
	if !ok {
		value = defaultValue
	}
	return fmt.Sprintf("%s=%s", m.Name, value(m))
}

// perCoreValue summarises per-core readings as "Nc(avgX%,maxY%)".
func perCoreValue(m domain.Metric) string {
	max, sum := m.PerCore[0], 0.0
	for _, v := range m.PerCore {
		if v > max {
			max = v
		}
		sum += v
	}
	return fmt.Sprintf("%dc(avg%.0f%%,max%.0f%%)", len(m.PerCore), sum/float64(len(m.PerCore)), max)
}

func shortDuration(secs float64) string {
	d := time.Duration(secs) * time.Second
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	m := int(d.Minutes()) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, h)
	case h > 0:
		return fmt.Sprintf("%dh%dm", h, m)
	default:
		return fmt.Sprintf("%dm", m)
	}
}
