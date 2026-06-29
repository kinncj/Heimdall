// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/transport"
	v1 "heimdall/common/proto/monitoring/v1"
)

// runCLI implements `heimdall-hub cli <command>` — a one-shot, machine-readable
// (JSON) view of a running hub's fleet, for scripts and AI agents. It subscribes
// to the hub like a dashboard, gathers the current state, prints JSON, and exits.
func runCLI(args []string) {
	fs := flag.NewFlagSet("cli", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	hubAddr := fs.String("hub", "localhost:9090", "hub address to query")
	token := fs.String("token", os.Getenv("HEIMDALL_TOKEN"), "enrollment token (env HEIMDALL_TOKEN)")
	tls := fs.Bool("tls", false, "connect to the hub over TLS")
	caFile := fs.String("tls-ca", "", "PEM CA bundle to trust")
	serverName := fs.String("tls-server-name", "", "override the verified server name")
	skipVerify := fs.Bool("tls-insecure", false, "skip hub certificate verification (dev only)")
	wait := fs.Duration("wait", 800*time.Millisecond, "how long to gather hub state before printing")
	fs.Usage = cliUsage
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	cmd := fs.Arg(0)
	if cmd == "" || cmd == "help" {
		cliUsage()
		if cmd == "" {
			os.Exit(2)
		}
		return
	}

	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	if err := cliGather(reg, *hubAddr, *token, secure.ClientConfig{
		Enabled: *tls, CAFile: *caFile, ServerName: *serverName, SkipVerify: *skipVerify,
	}, *wait); err != nil {
		cliFail(err)
	}
	if err := cliDispatch(reg, cmd, fs.Args()[1:]); err != nil {
		cliFail(err)
	}
}

func cliUsage() {
	fmt.Fprint(os.Stderr, `heimdall-hub cli — machine-readable fleet queries (JSON)

Usage:
  heimdall-hub cli [--hub addr] [--token t] [--tls …] <command> [args]

Commands:
  fleet                 fleet summary: host counts by state
  hosts                 every host with state, labels, metrics, capabilities
  host <id>             one host in full (metrics, processes, log sources)
  top <id>              the host's latest process table
  logs <id> [source]    the host's buffered log lines (optionally one source)

All output is JSON on stdout; errors are JSON on stderr with a non-zero exit.

Examples:
  heimdall-hub cli hosts | jq '.[] | select(.state=="offline").id'
  heimdall-hub cli top dgx-spark
  heimdall-hub cli logs web-01 app
`)
}

func cliFail(err error) {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	fmt.Fprintln(os.Stderr, string(b))
	os.Exit(1)
}

// cliGather subscribes to the hub and folds its current-state burst into reg over
// a short window, then computes liveness.
func cliGather(reg *domain.HostRegistry, addr, token string, tlsCfg secure.ClientConfig, wait time.Duration) error {
	transportOpt, err := tlsCfg.DialOption()
	if err != nil {
		return err
	}
	opts := []grpc.DialOption{transportOpt}
	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(secure.TokenCredentials{Token: token}))
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	stream, err := v1.NewFederationServiceClient(conn).Subscribe(ctx, &v1.SubscribeRequest{SubscriberId: "hub-cli"})
	if err != nil {
		return err
	}
	n := 0
	for {
		snap, err := stream.Recv()
		if err != nil {
			if n == 0 && ctx.Err() == nil { // a real failure before the window elapsed
				return fmt.Errorf("subscribe to %s: %w", addr, err)
			}
			break // window elapsed (deadline) or stream closed
		}
		cliFold(reg, snap)
		n++
	}
	reg.Evaluate(time.Now())
	return nil
}

func cliFold(reg *domain.HostRegistry, snap *v1.Snapshot) {
	id, ms, labels := transport.FromSnapshot(snap)
	hid := domain.HostID(id)
	seen := time.Now()
	if t := snap.GetTsUnixMillis(); t > 0 {
		seen = time.UnixMilli(t)
	}
	reg.Enroll(domain.Host{ID: hid, Hostname: id, DisplayName: id}, seen)
	reg.Observe(hid, ms, labels, seen)
	reg.SetAlerts(hid, snap.GetAlerts())
	procs, procsAt, logs := transport.ObservabilityFromSnapshot(snap)
	reg.RecordPush(hid, procs, procsAt, logs)
}

func cliDispatch(reg *domain.HostRegistry, cmd string, args []string) error {
	switch cmd {
	case "fleet":
		return cliEmit(fleetSummary(reg))
	case "hosts":
		hs := reg.Hosts()
		out := make([]jHost, len(hs))
		for i, h := range hs {
			out[i] = newJHost(h)
		}
		return cliEmit(out)
	case "host":
		h, err := cliHost(reg, args)
		if err != nil {
			return err
		}
		return cliEmit(newJHostDetail(h))
	case "top":
		h, err := cliHost(reg, args)
		if err != nil {
			return err
		}
		return cliEmit(newJTop(h))
	case "logs":
		if len(args) == 0 {
			return fmt.Errorf("logs needs a host id")
		}
		h, err := cliHost(reg, args[:1])
		if err != nil {
			return err
		}
		source := ""
		if len(args) > 1 {
			source = args[1]
		}
		return cliEmit(newJLogs(h, source))
	default:
		return fmt.Errorf("unknown command %q (try: fleet, hosts, host, top, logs)", cmd)
	}
}

func cliHost(reg *domain.HostRegistry, args []string) (domain.HostView, error) {
	if len(args) == 0 {
		return domain.HostView{}, fmt.Errorf("a host id is required")
	}
	h, ok := reg.Host(domain.HostID(args[0]))
	if !ok {
		return domain.HostView{}, fmt.Errorf("no such host %q", args[0])
	}
	return h, nil
}

func cliEmit(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// ---- JSON shapes (snake_case, stable for agent consumption) ----------------

type jHost struct {
	ID           string             `json:"id"`
	State        string             `json:"state"`
	LastSeenUnix int64              `json:"last_seen_unix"`
	Labels       map[string]string  `json:"labels,omitempty"`
	Metrics      map[string]float64 `json:"metrics,omitempty"`
	Alerts       []string           `json:"alerts,omitempty"`
	HasLogs      bool               `json:"has_logs"`
	HasProcesses bool               `json:"has_processes"`
	LogSources   []string           `json:"log_sources,omitempty"`
}

type jProc struct {
	PID     uint32  `json:"pid"`
	PPID    uint32  `json:"ppid"`
	CPUPct  float64 `json:"cpu_pct"`
	MemPct  float64 `json:"mem_pct"`
	Command string  `json:"command"`
}

type jLine struct {
	Source      string `json:"source"`
	TSUnix      int64  `json:"ts_unix"`
	Line        string `json:"line"`
	Level       string `json:"level,omitempty"`
	RateLimited bool   `json:"rate_limited,omitempty"`
}

type jHostDetail struct {
	jHost
	ProcessesAtUnix int64   `json:"processes_at_unix,omitempty"`
	Processes       []jProc `json:"processes,omitempty"`
}

type jTop struct {
	Host            string  `json:"host"`
	ProcessesAtUnix int64   `json:"processes_at_unix"`
	Processes       []jProc `json:"processes"`
}

type jLogs struct {
	Host   string  `json:"host"`
	Source string  `json:"source,omitempty"`
	Lines  []jLine `json:"lines"`
}

// userLabels strips reserved (hub/daemon-managed, "_"-prefixed) keys.
func userLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	out := make(map[string]string, len(labels))
	for k, v := range labels {
		if !strings.HasPrefix(k, "_") {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func newJHost(h domain.HostView) jHost {
	metrics := map[string]float64{}
	for _, m := range h.LastSnapshot {
		if m.Status == domain.StatusOK {
			metrics[m.Name] = m.Gauge
		}
	}
	var sources []string
	if v := h.Host.Context.Labels["_logs"]; v != "" {
		sources = strings.Split(v, ",")
	}
	return jHost{
		ID:           string(h.Host.ID),
		State:        h.State.String(),
		LastSeenUnix: h.LastSeen.Unix(),
		Labels:       userLabels(h.Host.Context.Labels),
		Metrics:      metrics,
		Alerts:       h.Alerts,
		HasLogs:      len(sources) > 0,
		HasProcesses: h.Host.Context.Labels["_proc"] != "" || len(h.Processes) > 0,
		LogSources:   sources,
	}
}

func newJProcs(rows []domain.ProcessRow) []jProc {
	out := make([]jProc, len(rows))
	for i, p := range rows {
		out[i] = jProc{PID: p.PID, PPID: p.PPID, CPUPct: p.CPUPct, MemPct: p.MemPct, Command: p.Command}
	}
	return out
}

func newJHostDetail(h domain.HostView) jHostDetail {
	d := jHostDetail{jHost: newJHost(h), Processes: newJProcs(h.Processes)}
	if !h.ProcessesAt.IsZero() {
		d.ProcessesAtUnix = h.ProcessesAt.Unix()
	}
	return d
}

func newJTop(h domain.HostView) jTop {
	t := jTop{Host: string(h.Host.ID), Processes: newJProcs(h.Processes)}
	if !h.ProcessesAt.IsZero() {
		t.ProcessesAtUnix = h.ProcessesAt.Unix()
	}
	return t
}

func newJLogs(h domain.HostView, source string) jLogs {
	lines := []jLine{}
	for _, l := range h.Logs {
		if source != "" && l.Source != source {
			continue
		}
		lines = append(lines, jLine{
			Source: l.Source, TSUnix: l.At.Unix(), Line: l.Line,
			Level: l.Level, RateLimited: l.RateLimited,
		})
	}
	return jLogs{Host: string(h.Host.ID), Source: source, Lines: lines}
}

func fleetSummary(reg *domain.HostRegistry) map[string]any {
	hosts := reg.Hosts()
	by := map[string]int{}
	ids := make([]string, 0, len(hosts))
	for _, h := range hosts {
		by[h.State.String()]++
		ids = append(ids, string(h.Host.ID))
	}
	sort.Strings(ids)
	return map[string]any{
		"total":    len(hosts),
		"by_state": by,
		"host_ids": ids,
		"gathered": time.Now().Unix(),
	}
}
