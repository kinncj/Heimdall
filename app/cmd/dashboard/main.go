// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Command heimdall-dashboard renders the Heimdall TUI. With --snapshot it prints
// a single frame to stdout (no TTY needed); otherwise it runs interactively.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
	"google.golang.org/grpc"

	"heimdall/app/internal/discovery"
	"heimdall/app/internal/domain"
	"heimdall/app/internal/fake"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/selfupdate"
	"heimdall/app/internal/transport"
	"heimdall/app/internal/tui/dashboard"
	"heimdall/app/internal/tui/splash"
	"heimdall/app/internal/tui/theme"
	v1 "heimdall/common/proto/monitoring/v1"
)

// version is the Heimdall build version, set via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := selfupdate.Run("dashboard", version); err != nil {
			fmt.Fprintln(os.Stderr, "heimdall-dashboard:", err)
			os.Exit(1)
		}
		return
	}
	snapshot := flag.Bool("snapshot", false, "render one grid frame to stdout and exit")
	splashFlag := flag.Bool("splash", false, "render the splash frame to stdout and exit")
	demoMode := flag.Bool("demo", false, "render a simulated multi-host fleet (no hub needed; for trying the UI)")
	detailFlag := flag.Bool("detail", false, "render the host-detail frame (with --snapshot)")
	showVersion := flag.Bool("version", false, "print version and exit")
	cat := dashboardCatalog()
	cat.Register(flag.CommandLine)
	flag.Usage = usage
	flag.Parse()
	if *showVersion {
		fmt.Println("heimdall-dashboard", version)
		return
	}

	cfg := resolveDashboard(cat)
	token := cfg.Secret("token").Reveal()
	tlsCfg := secure.ClientConfig{
		Enabled: cfg.Toggle("tls"), CAFile: cfg.Text("tls-ca"),
		ServerName: cfg.Text("tls-server-name"), SkipVerify: cfg.Toggle("tls-insecure"),
	}
	hubAddr := cfg.Address("hub").String()

	th, err := theme.Load()
	if err != nil {
		fail(err)
	}
	md, ok := th.Mode(cfg.Text("mode"))
	if !ok {
		fail(fmt.Errorf("unknown theme mode %q", cfg.Text("mode")))
	}

	now := time.Now()
	if *splashFlag {
		fmt.Println(splash.Render(md, 100, 34))
		return
	}

	// The dashboard is a pure presentation surface: it renders host metrics that
	// daemons publish to a hub. It never collects this machine's metrics itself
	// (run a daemon for that — it can live on the same machine). --demo stands in
	// a synthetic fleet so the UI can be explored without any infrastructure.
	var reg *domain.HostRegistry
	var tickFn func(time.Time)
	var source string
	var live func() bool
	var runCmd func(host, cmd string, args []string, reqID string) // v2 Phase 2; nil in demo
	switch {
	case *demoMode:
		src := fake.New(now)
		reg = src.Registry()
		tickFn = src.Tick
		source = "demo"
		live = func() bool { return true }
		// Fake command execution so the command modal is explorable in --demo.
		runCmd = func(host, cmd string, args []string, reqID string) {
			reg.RecordCommandResult(domain.HostID(host), &domain.CommandResult{
				RequestID: reqID, Status: domain.StatusOK,
				Stdout: fmt.Sprintf("(demo) %s on %s\nconnect to a real hub to run live commands\n", cmd, host),
			})
		}
	default:
		reg = domain.NewHostRegistry(10*time.Second, 30*time.Second)
		reg.SetPurgeAfter(cfg.Span("purge-after", 15*time.Minute))
		// Ratatoskr: discover the hub when --hub is "auto" (or --discover with no
		// explicit hub). Discovery only resolves the address; the token and TLS
		// still gate trust. An explicit --hub always wins. When several hubs are
		// found on the LAN, present a picker so the operator chooses (v2).
		if cfg.Text("hub") == "auto" || (cfg.Toggle("discover") && cfg.Text("hub") == "") {
			hubs, _ := discovery.BrowseAll(3 * time.Second)
			switch len(hubs) {
			case 1:
				hubAddr = hubs[0].Addr
			case 0:
				// fall back to a static seed for overlay networks (no multicast)
				addr, err := discovery.Resolve(cfg.Text("discover-seed"), 5*time.Second)
				if err != nil {
					fail(fmt.Errorf("hub discovery failed (Ratatoskr): %w", err))
				}
				hubAddr = addr
			default:
				chosen, err := pickHub(hubs, md)
				if err != nil {
					fail(err)
				}
				hubAddr = chosen
			}
		}
		dialOpts, err := clientDialOptions(token, tlsCfg)
		if err != nil {
			fail(err)
		}
		lastRecv := new(atomic.Int64)
		go subscribeHub(hubAddr, reg, dialOpts, lastRecv)
		// v2 Phase 2: issue on-demand commands to the hub. Fire-and-forget — the
		// result returns on the subscription and lands in the registry; the modal
		// shows it. A dedicated connection keeps it off the subscription stream.
		if cmdConn, err := grpc.NewClient(hubAddr, dialOpts...); err == nil {
			fed := v1.NewFederationServiceClient(cmdConn)
			actor := os.Getenv("USER")
			if actor == "" {
				actor = "dashboard"
			}
			runCmd = func(host, cmd string, args []string, reqID string) {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					_, _ = fed.RunCommand(ctx, &v1.ControlRequest{
						RequestId: reqID, HostId: host, AllowlistedCmd: cmd, Args: args, Actor: actor,
					})
				}()
			}
		}
		source = hubAddr
		live = func() bool {
			ms := lastRecv.Load()
			return ms > 0 && time.Since(time.UnixMilli(ms)) < 5*time.Second
		}
		if *snapshot {
			time.Sleep(1200 * time.Millisecond) // gather a few updates for the frame
		}
	}

	model := dashboard.New(md, reg, now).WithStatus(source, live).
		WithTopSort(cfg.Text("top-sort")).WithPersistSort(saveTopSort)
	if runCmd != nil {
		model = model.WithRunCommand(runCmd)
	}
	if tickFn != nil {
		model = model.WithTick(tickFn)
	}

	if *snapshot {
		reg.Evaluate(time.Now())
		if *detailFlag {
			fmt.Println(model.DetailView())
		} else {
			fmt.Println(model.GridView())
		}
		return
	}

	// Brand splash: rendered ONCE on the primary screen (an inline image on
	// capable terminals, else clean ASCII), held briefly, then cleared before
	// Bubble Tea takes the alternate screen — so it neither double-renders nor
	// lingers behind the grid.
	sw, sh := 100, 34
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		sw, sh = w, h
	}
	fmt.Print("\x1b[2J\x1b[H") // blank screen so the splash centres cleanly
	fmt.Print(splash.Render(md, sw, sh))
	time.Sleep(1500 * time.Millisecond)
	fmt.Print("\x1b[2J\x1b[3J\x1b[H")

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "heimdall-dashboard:", err)
	os.Exit(1)
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, `heimdall-dashboard — Heimdall monitoring dashboard (presentation only)

The dashboard renders host metrics that daemons publish to a hub. It does not
collect metrics itself; to monitor a machine, run heimdall-daemon there (it may
be the same machine as the dashboard).

Usage:
  heimdall-dashboard [flags]

Common:
  heimdall-dashboard --hub localhost:9090         subscribe to a hub (default)
  heimdall-dashboard --demo                       explore the UI with a synthetic fleet
  heimdall-dashboard --hub HOST:9090              logs (l) and top (t) live in the host detail view

Flags:
`)
	flag.PrintDefaults()
}

// clientDialOptions assembles the transport security and per-RPC enrollment
// token used to reach the hub.
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

// subscribeHub keeps a live subscription to the hub, folding incoming snapshots
// into the registry; it reconnects on error so the dashboard self-heals. Each
// folded snapshot stamps lastRecv so the footer can report live connectivity.
func subscribeHub(addr string, reg *domain.HostRegistry, dialOpts []grpc.DialOption, lastRecv *atomic.Int64) {
	for {
		conn, err := grpc.NewClient(addr, dialOpts...)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		stream, err := v1.NewFederationServiceClient(conn).Subscribe(
			context.Background(), &v1.SubscribeRequest{SubscriberId: "dashboard"})
		if err != nil {
			_ = conn.Close()
			time.Sleep(time.Second)
			continue
		}
		for {
			snap, err := stream.Recv()
			if err != nil {
				break
			}
			foldSnapshot(reg, snap)
			lastRecv.Store(time.Now().UnixMilli())
		}
		_ = conn.Close()
		time.Sleep(time.Second)
	}
}

// foldSnapshot folds a hub snapshot into the registry, observing the host at the
// snapshot's own timestamp (the hub's last-seen time for that host) rather than
// wall-clock now. This is what lets a host go OFFLINE: a retained snapshot for a
// dead daemon — replayed on connect or reconnect — ages from its real timestamp
// instead of looking freshly seen. A missing timestamp falls back to now.
func foldSnapshot(reg *domain.HostRegistry, snap *v1.Snapshot) {
	id, ms, labels := transport.FromSnapshot(snap)
	hid := domain.HostID(id)
	seen := time.Now()
	if t := snap.GetTsUnixMillis(); t > 0 {
		seen = time.UnixMilli(t)
	}
	reg.Enroll(domain.Host{ID: hid, Hostname: id, DisplayName: id}, seen)
	reg.Observe(hid, ms, labels, seen)
	reg.SetAlerts(hid, snap.GetAlerts())
	// Heimdallr's sight (ADR 0017): fold the host's pushed process table and any
	// tailed log lines into the registry for the detail-view modals.
	procs, procsAt, logLines := transport.ObservabilityFromSnapshot(snap)
	reg.RecordPush(hid, procs, procsAt, logLines)
	reg.RecordCommandResult(hid, transport.CommandResultFromSnapshot(snap)) // v2 Phase 2
}
