// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Command heimdall-daemon samples this host's metrics through the unprivileged
// adapters. With --hub it streams them to a hub over the versioned gRPC contract
// (auto-reconnecting); without it, it prints samples (text or --json).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"heimdall/app/internal/adapters"
	"heimdall/app/internal/control"
	"heimdall/app/internal/domain"
	"heimdall/app/internal/logs"
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
	hubAddr := flag.String("hub", "", "hub address to stream to (e.g. localhost:9090); empty prints locally")
	name := flag.String("name", "", "host display name (default: hostname)")
	interval := flag.Duration("interval", 2*time.Second, "sample interval")
	once := flag.Bool("once", false, "collect a single sample and exit (print mode)")
	asJSON := flag.Bool("json", false, "emit one JSON object per metric (print mode)")
	token := flag.String("token", os.Getenv("HEIMDALL_TOKEN"), "enrollment token presented to the hub (env HEIMDALL_TOKEN)")
	useTLS := flag.Bool("tls", false, "connect to the hub over TLS")
	tlsCA := flag.String("tls-ca", "", "PEM CA bundle to trust (default: system roots)")
	tlsServerName := flag.String("tls-server-name", "", "override the server name verified in the hub certificate")
	tlsInsecure := flag.Bool("tls-insecure", false, "skip hub certificate verification (dev only)")
	controlListen := flag.String("control-listen", "", "serve the read-only control plane on this address (e.g. :9100)")
	controlToken := flag.String("control-token", os.Getenv("HEIMDALL_CONTROL_TOKEN"), "token required to invoke control commands (env HEIMDALL_CONTROL_TOKEN)")
	controlTLSCert := flag.String("control-tls-cert", "", "PEM server cert for the control plane (enables TLS with --control-tls-key)")
	controlTLSKey := flag.String("control-tls-key", "", "PEM server key for the control plane")
	logSource := flag.String("log-source", "", "opt-in log sources alias=path,alias2=path2 (served on --control-listen; empty = logs off)")
	logFile := flag.String("log-file", "", "operational log destination: unset = stderr (TTY); 'false' = disabled; a path = JSON logs appended to that file")
	pingTarget := flag.String("ping-target", "1.1.1.1", "internet host pinged for reachability/latency (net.latency)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()
	if *showVersion {
		fmt.Println("heimdall-daemon", version)
		return
	}

	lg, metricsW, metricsJSON, closeLog, err := resolveOutput(*logFile, *asJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, "heimdall-daemon:", err)
		os.Exit(1)
	}
	defer closeLog()
	logger = lg

	host := *name
	if host == "" {
		if h, err := os.Hostname(); err == nil && h != "" {
			host = h
		} else {
			host = "localhost"
		}
	}

	reg := domain.NewRegistry(*interval)
	for _, a := range adapters.Build(adapters.Options{PingTarget: *pingTarget, Version: version}) {
		reg.Register(a)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	if *controlListen != "" {
		sources, err := logs.ParseSources(*logSource)
		if err != nil {
			fatal("invalid --log-source", err)
		}
		if err := startControlServer(*controlListen, *controlToken, *controlTLSCert, *controlTLSKey, host, sources); err != nil {
			fatal("control plane failed to start", err)
		}
	}

	if *hubAddr != "" {
		dialOpts, err := clientDialOptions(*token, secure.ClientConfig{
			Enabled: *useTLS, CAFile: *tlsCA, ServerName: *tlsServerName, SkipVerify: *tlsInsecure,
		})
		if err != nil {
			fatal("invalid client TLS configuration", err)
		}
		streamToHub(*hubAddr, host, *interval, reg, sig, dialOpts)
		return
	}

	sample := func() { emit(metricsW, host, reg.Collect(context.Background()), metricsJSON) }
	if *once {
		sample()
		return
	}
	logger.Info("daemon started",
		"host", host, "os", runtime.GOOS, "arch", runtime.GOARCH, "interval", interval.String())
	t := time.NewTicker(*interval)
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
  heimdall-daemon --control-listen :9100 --control-token "$T" --log-source app=/var/log/app.log
  heimdall-daemon --log-file /var/log/heimdall/daemon.json

Flags:
`)
	flag.PrintDefaults()
}

// auditLogger adapts the control-plane audit trail onto the structured logger so
// every invocation is one JSON event alongside the daemon's other logs.
type auditLogger struct{ log *slog.Logger }

func (a auditLogger) Record(e control.AuditEntry) {
	decision := "allow"
	if !e.Allowed {
		decision = "refuse"
	}
	a.log.Info("control audit",
		"actor", e.Actor,
		"command", e.Command,
		"args", e.Args,
		"decision", decision,
		"exit_code", e.ExitCode,
	)
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

// startControlServer serves the read-only control plane and (when sources are
// configured) the opt-in log stream on addr, optionally over TLS and behind a
// token. Both run as this unprivileged daemon user, on distinct gRPC services.
func startControlServer(addr, token, tlsCert, tlsKey, host string, sources logs.Sources) error {
	creds, err := secure.ServerOption(tlsCert, tlsKey)
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	auth := secure.NewAuthenticator(token)
	srv := grpc.NewServer(
		creds,
		grpc.UnaryInterceptor(secure.UnaryServerInterceptor(auth)),
		grpc.StreamInterceptor(secure.StreamServerInterceptor(auth)),
	)
	exec := control.NewExecutor(auditLogger{log: logger})
	v1.RegisterControlPlaneServiceServer(srv, control.NewServer(exec, host))
	v1.RegisterLogStreamServiceServer(srv, logs.NewServer(sources, host))
	logger.Info("control plane serving",
		"addr", addr, "tls", tlsCert != "", "auth", token != "", "log_sources", sources.Aliases())
	go func() {
		if err := srv.Serve(lis); err != nil {
			logger.Error("control server stopped", "error", err.Error())
		}
	}()
	return nil
}

// streamToHub streams snapshots to the hub, reconnecting with backoff on error.
func streamToHub(addr, host string, interval time.Duration, reg *domain.Registry, sig chan os.Signal, dialOpts []grpc.DialOption) {
	backoff := time.Second
	for {
		connected, err := streamOnce(addr, host, interval, reg, sig, dialOpts)
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

func streamOnce(addr, host string, interval time.Duration, reg *domain.Registry, sig chan os.Signal, dialOpts []grpc.DialOption) (bool, error) {
	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return false, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	ctx := context.Background()
	enrollReq := &v1.EnrollRequest{Host: &v1.Host{
		HostId: host, Hostname: host, DisplayName: host,
		Context: &v1.HostContext{Os: runtime.GOOS, Arch: runtime.GOARCH},
	}}
	if _, err := v1.NewEnrollmentServiceClient(conn).Enroll(ctx, enrollReq); err != nil {
		return false, fmt.Errorf("enroll: %w", err)
	}
	stream, err := v1.NewMetricStreamServiceClient(conn).Stream(ctx)
	if err != nil {
		return false, fmt.Errorf("open stream: %w", err)
	}
	logger.Info("streaming to hub", "addr", addr, "host", host)

	var seq uint64
	send := func() error {
		seq++
		return stream.Send(transport.ToSnapshot(host, reg.Collect(ctx), seq, time.Now()))
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
			Cores  []float64 `json:"cores,omitempty"`
		}
		enc := json.NewEncoder(w)
		now := time.Now().Unix()
		for _, m := range ms {
			_ = enc.Encode(row{host, now, m.Name, m.Status.String(), m.Gauge, m.Unit, m.PerCore})
		}
		return
	}
	line := host
	for _, m := range ms {
		line += "  " + formatMetric(m)
	}
	fmt.Fprintln(w, line)
}

// formatMetric renders one metric for the human-readable print line, handling
// per-core metrics and unit-appropriate precision.
func formatMetric(m domain.Metric) string {
	if m.Status != domain.StatusOK {
		return fmt.Sprintf("%s=%s", m.Name, m.Status)
	}
	if m.Kind == domain.KindPerCore && len(m.PerCore) > 0 {
		max, sum := m.PerCore[0], 0.0
		for _, v := range m.PerCore {
			if v > max {
				max = v
			}
			sum += v
		}
		return fmt.Sprintf("%s=%dc(avg%.0f%%,max%.0f%%)", m.Name, len(m.PerCore), sum/float64(len(m.PerCore)), max)
	}
	switch m.Unit {
	case "percent":
		return fmt.Sprintf("%s=%.0f%%", m.Name, m.Gauge)
	case "watts":
		return fmt.Sprintf("%s=%.2fW", m.Name, m.Gauge)
	case "celsius":
		return fmt.Sprintf("%s=%.0f°C", m.Name, m.Gauge)
	case "ms":
		return fmt.Sprintf("%s=%.0fms", m.Name, m.Gauge)
	case "MB/s":
		return fmt.Sprintf("%s=%.2fMB/s", m.Name, m.Gauge)
	case "s":
		return fmt.Sprintf("%s=%s", m.Name, shortDuration(m.Gauge))
	default:
		return fmt.Sprintf("%s=%g", m.Name, m.Gauge)
	}
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
