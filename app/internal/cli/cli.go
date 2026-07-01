// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"heimdall/app/internal/command"
	"heimdall/app/internal/discovery"
	"heimdall/app/internal/domain"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/transport"
	v1 "heimdall/common/proto/monitoring/v1"
)

// Run executes the fleet CLI (`heimdall-cli <command>`, also reachable as the
// `cli` subcommand of `heimdall-hub`): a one-shot, machine-readable (JSON) view of
// a running hub's fleet. It subscribes like a dashboard, gathers the current
// state, prints JSON, and exits — for scripts and AI agents.
func Run(args []string) {
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

	tlsCfg := secure.ClientConfig{
		Enabled: *tls, CAFile: *caFile, ServerName: *serverName, SkipVerify: *skipVerify,
	}

	// --hub auto browses the LAN (Ratatoskr). One hub → use it; several → report
	// them and tell the operator to pick with --hub, exactly as asked; none → error.
	hub := *hubAddr
	if hub == "auto" {
		hubs, err := discovery.BrowseAll(3 * time.Second)
		if err != nil {
			cliFail(err)
		}
		switch len(hubs) {
		case 0:
			cliFail(fmt.Errorf("no hub discovered on the LAN; set --hub <addr>"))
		case 1:
			hub = hubs[0].Addr
		default:
			cliReportHubs(hubs) // prints the list + instruction, exits non-zero
		}
	}

	// `run` issues an on-demand command and waits for its result; it does not fit
	// the gather-then-print model the read-only commands use.
	if cmd == "run" {
		if err := cliRun(hub, *token, tlsCfg, fs.Args()[1:], *wait); err != nil {
			cliFail(err)
		}
		return
	}

	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	if err := cliGather(reg, hub, *token, tlsCfg, *wait); err != nil {
		cliFail(err)
	}
	if err := cliDispatch(reg, cmd, fs.Args()[1:]); err != nil {
		cliFail(err)
	}
}

func cliUsage() {
	fmt.Fprint(os.Stderr, `heimdall-cli — machine-readable fleet queries (JSON)

Usage:
  heimdall-cli [--hub addr|auto] [--token t] [--tls …] <command> [args]

  --hub auto discovers the hub over mDNS (Ratatoskr). If several hubs are found it
  lists them and asks you to pick one with --hub <addr>.

Commands:
  fleet                 fleet summary: host counts by state
  hosts                 every host with state, labels, metrics, capabilities
  host <id>             one host in full (metrics, processes, log sources)
  top <id>              the host's latest process table
  logs <id> [source]    the host's buffered log lines (optionally one source)
  run <id> <cmd> [args] run an allow-listed read-only command on a host (v2)
                        cmd: process.list | disk.df | uptime | os.info | dir.list <dir>
                        privileged (need the root helper): dmesg | journal.tail

All output is JSON on stdout; errors are JSON on stderr with a non-zero exit.

Examples:
  heimdall-cli hosts | jq '.[] | select(.state=="offline").id'
  heimdall-cli top dgx-spark
  heimdall-cli logs web-01 app
  heimdall-cli run web-01 disk.df | jq -r .stdout
`)
}

func cliFail(err error) {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	fmt.Fprintln(os.Stderr, string(b))
	os.Exit(1)
}

// cliReportHubs prints the discovered hubs and tells the operator to pick one,
// then exits non-zero. Only reached when --hub auto finds more than one (per the
// requirement to report-and-instruct only on multiple zeroconf hubs).
func cliReportHubs(hubs []discovery.Found) {
	type jHub struct {
		Name string `json:"name"`
		Addr string `json:"addr"`
	}
	list := make([]jHub, len(hubs))
	for i, h := range hubs {
		list[i] = jHub{Name: h.Name, Addr: h.Addr}
	}
	b, _ := json.MarshalIndent(map[string]any{
		"error": "multiple hubs discovered — choose one with --hub <addr>",
		"hubs":  list,
	}, "", "  ")
	fmt.Fprintln(os.Stderr, string(b))
	os.Exit(1)
}

// cliDial opens a hub client connection with the configured transport security
// and per-RPC enrollment token.
func cliDial(addr, token string, tlsCfg secure.ClientConfig) (*grpc.ClientConn, error) {
	transportOpt, err := tlsCfg.DialOption()
	if err != nil {
		return nil, err
	}
	opts := []grpc.DialOption{transportOpt}
	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(secure.TokenCredentials{Token: token}))
	}
	return grpc.NewClient(addr, opts...)
}

// cliRun issues an on-demand allow-listed command via the hub and waits for its
// result (v2 Phase 2, ADR 0018). It subscribes first so it cannot miss the result
// snapshot, then issues the command and reads until the matching result arrives.
func cliRun(addr, token string, tlsCfg secure.ClientConfig, args []string, wait time.Duration) error {
	if len(args) < 1 {
		return fmt.Errorf("run needs: <host> <command> [args] (commands: %s)", strings.Join(command.Keys(), ", "))
	}
	host := args[0]
	if len(args) < 2 {
		return fmt.Errorf("run needs a command (one of: %s)", strings.Join(command.Keys(), ", "))
	}
	cmdKey, cmdArgs := args[1], args[2:]

	conn, err := cliDial(addr, token, tlsCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	fed := v1.NewFederationServiceClient(conn)

	timeout := wait
	if timeout < 12*time.Second {
		timeout = 12 * time.Second // a command runs longer than a state gather
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stream, err := fed.Subscribe(ctx, &v1.SubscribeRequest{SubscriberId: "hub-cli-run"})
	if err != nil {
		return err
	}
	actor := os.Getenv("USER")
	if actor == "" {
		actor = "heimdall-cli"
	}
	reqID := fmt.Sprintf("cli-%d", time.Now().UnixNano())
	ack, err := fed.RunCommand(ctx, &v1.ControlRequest{
		RequestId: reqID, HostId: host, AllowlistedCmd: cmdKey, Args: cmdArgs, Actor: actor,
	})
	if err != nil {
		return err
	}
	if !ack.GetAccepted() {
		return fmt.Errorf("%s", ack.GetError())
	}
	for {
		snap, err := stream.Recv()
		if err != nil {
			// Only a deadline is really a timeout; surface anything else (permission
			// denied, connection reset, hub shutdown) as itself rather than masking it.
			if errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded {
				return fmt.Errorf("timed out waiting for the result of %q on %q", cmdKey, host)
			}
			return fmt.Errorf("stream error waiting for the result of %q on %q: %w", cmdKey, host, err)
		}
		if cr := snap.GetCommandResult(); cr != nil && cr.GetRequestId() == reqID {
			return cliEmit(newJCommand(host, cmdKey, cr))
		}
	}
}

type jCommand struct {
	Host      string `json:"host"`
	Command   string `json:"command"`
	Status    string `json:"status"`
	ExitCode  int32  `json:"exit_code"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

func newJCommand(host, cmd string, cr *v1.ControlResponse) jCommand {
	return jCommand{
		Host: host, Command: cmd,
		Status:    statusString(cr.GetStatus()),
		ExitCode:  cr.GetExitCode(),
		Stdout:    cr.GetStdout(),
		Stderr:    cr.GetStderr(),
		Truncated: cr.GetTruncated(),
	}
}

func statusString(s v1.MetricStatus) string {
	switch s {
	case v1.MetricStatus_METRIC_STATUS_OK:
		return "ok"
	case v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION:
		return "insufficient_permission"
	case v1.MetricStatus_METRIC_STATUS_ERROR:
		return "error"
	default:
		return "unspecified"
	}
}

// cliGather subscribes to the hub and folds its current-state burst into reg over
// a short window, then computes liveness.
func cliGather(reg *domain.HostRegistry, addr, token string, tlsCfg secure.ClientConfig, wait time.Duration) error {
	conn, err := cliDial(addr, token, tlsCfg)
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
	ID           string                   `json:"id"`
	State        string                   `json:"state"`
	LastSeenUnix int64                    `json:"last_seen_unix"`
	Labels       map[string]string        `json:"labels,omitempty"`
	Metrics      map[string]float64       `json:"metrics,omitempty"`
	Details      map[string]string        `json:"details,omitempty"`
	Unavailable  map[string]jMetricStatus `json:"unavailable,omitempty"`
	Alerts       []string                 `json:"alerts,omitempty"`
	HasLogs      bool                     `json:"has_logs"`
	HasProcesses bool                     `json:"has_processes"`
	LogSources   []string                 `json:"log_sources,omitempty"`
}

// jMetricStatus explains a non-OK metric — why a value is missing (e.g. Apple
// gpu.vram "unified memory (no discrete VRAM)", npu.util "no NPU counter", or a
// metric that needs the privileged helper). OK metrics stay in Metrics as bare
// values; this keeps that map backward-compatible for existing scripts.
type jMetricStatus struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
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
	var details map[string]string
	var unavailable map[string]jMetricStatus
	for _, m := range h.LastSnapshot {
		if m.Status == domain.StatusOK {
			metrics[m.Name] = m.Gauge
			// Info metrics (host.version, host.gpu, …) carry their value in Detail
			// with no gauge, and gauges like gpu.vram carry a used/total note. Expose
			// those strings so scripts can read them, not just the TUI.
			if m.Detail != "" {
				if details == nil {
					details = map[string]string{}
				}
				details[m.Name] = m.Detail
			}
			continue
		}
		if unavailable == nil {
			unavailable = map[string]jMetricStatus{}
		}
		unavailable[m.Name] = jMetricStatus{Status: m.Status.String(), Detail: m.Detail}
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
		Details:      details,
		Unavailable:  unavailable,
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
