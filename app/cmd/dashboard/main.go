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
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
	"google.golang.org/grpc"

	"heimdall/app/internal/control"
	"heimdall/app/internal/domain"
	"heimdall/app/internal/fake"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/transport"
	"heimdall/app/internal/tui/dashboard"
	"heimdall/app/internal/tui/splash"
	"heimdall/app/internal/tui/theme"
	v1 "heimdall/common/proto/monitoring/v1"
)

func main() {
	snapshot := flag.Bool("snapshot", false, "render one grid frame to stdout and exit")
	splashFlag := flag.Bool("splash", false, "render the splash frame to stdout and exit")
	demoMode := flag.Bool("demo", false, "render a simulated multi-host fleet (no hub needed; for trying the UI)")
	hubAddr := flag.String("hub", "localhost:9090", "hub address to subscribe to for live host metrics")
	detailFlag := flag.Bool("detail", false, "render the host-detail frame (with --snapshot)")
	mode := flag.String("mode", "dark", "theme mode: dark|light")
	token := flag.String("token", os.Getenv("HEIMDALL_TOKEN"), "enrollment token presented to the hub (env HEIMDALL_TOKEN)")
	useTLS := flag.Bool("tls", false, "connect to the hub over TLS")
	tlsCA := flag.String("tls-ca", "", "PEM CA bundle to trust (default: system roots)")
	tlsServerName := flag.String("tls-server-name", "", "override the server name verified in the hub certificate")
	tlsInsecure := flag.Bool("tls-insecure", false, "skip hub certificate verification (dev only)")
	controlAddr := flag.String("control", "", "daemon control-plane address for --run")
	controlRun := flag.String("run", "", "allow-listed control command to run against --control, e.g. \"process.list\" or \"dir.list /var/log\"")
	tailAlias := flag.String("tail", "", "tail an opt-in log source alias from --control (e.g. app); streams until ctrl-c")
	flag.Usage = usage
	flag.Parse()

	if *controlRun != "" || *tailAlias != "" {
		if *controlAddr == "" {
			fail(fmt.Errorf("--run/--tail require --control <addr>"))
		}
		dialOpts, err := clientDialOptions(*token, secure.ClientConfig{
			Enabled: *useTLS, CAFile: *tlsCA, ServerName: *tlsServerName, SkipVerify: *tlsInsecure,
		})
		if err != nil {
			fail(err)
		}
		if *controlRun != "" {
			runControl(*controlAddr, *controlRun, dialOpts)
		} else {
			runTail(*controlAddr, *tailAlias, dialOpts)
		}
		return
	}

	th, err := theme.Load()
	if err != nil {
		fail(err)
	}
	md, ok := th.Mode(*mode)
	if !ok {
		fail(fmt.Errorf("unknown theme mode %q", *mode))
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
	switch {
	case *demoMode:
		src := fake.New(now)
		reg = src.Registry()
		tickFn = src.Tick
	default:
		reg = domain.NewHostRegistry(10*time.Second, 30*time.Second)
		dialOpts, err := clientDialOptions(*token, secure.ClientConfig{
			Enabled: *useTLS, CAFile: *tlsCA, ServerName: *tlsServerName, SkipVerify: *tlsInsecure,
		})
		if err != nil {
			fail(err)
		}
		go subscribeHub(*hubAddr, reg, dialOpts)
		if *snapshot {
			time.Sleep(1200 * time.Millisecond) // gather a few updates for the frame
		}
	}

	model := dashboard.New(md, reg, now)
	if tickFn != nil {
		model = model.WithTick(tickFn)
	}

	if *snapshot {
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
  heimdall-dashboard --control HOST:PORT --run process.list
  heimdall-dashboard --control HOST:PORT --tail app

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

// runControl invokes one allow-listed control command against a daemon's
// control endpoint and prints the result (the "shown in the dashboard" path).
func runControl(addr, runSpec string, dialOpts []grpc.DialOption) {
	fields := strings.Fields(runSpec)
	if len(fields) == 0 {
		fail(fmt.Errorf("empty --run command"))
	}
	cmd, args := fields[0], fields[1:]

	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		fail(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	actor := os.Getenv("USER")
	if actor == "" {
		actor = "dashboard"
	}
	resp, err := (control.Client{Conn: conn}).Run(ctx, "", cmd, args, actor)
	if err != nil {
		fail(err)
	}
	if resp.GetStatus() != v1.MetricStatus_METRIC_STATUS_OK {
		fmt.Fprintf(os.Stderr, "heimdall-dashboard: refused (%s): %s\n", resp.GetStatus(), resp.GetStderr())
		os.Exit(1)
	}
	fmt.Print(resp.GetStdout())
	if resp.GetTruncated() {
		fmt.Fprintln(os.Stderr, "[output truncated]")
	}
}

// runTail streams an opt-in log source from a daemon's log endpoint to stdout
// until interrupted (the "logs pane" path, in line form).
func runTail(addr, alias string, dialOpts []grpc.DialOption) {
	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		fail(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	stream, err := v1.NewLogStreamServiceClient(conn).Tail(ctx, &v1.LogTailRequest{Sources: []string{alias}})
	if err != nil {
		fail(err)
	}
	for {
		line, err := stream.Recv()
		if err != nil {
			return
		}
		ts := time.UnixMilli(line.GetTsUnixMillis()).Format("15:04:05")
		marker := ""
		if line.GetRateLimited() {
			marker = " [rate-limited]"
		}
		fmt.Printf("%s %s%s  %s\n", ts, line.GetSource(), marker, line.GetLine())
	}
}

// subscribeHub keeps a live subscription to the hub, folding incoming snapshots
// into the registry; it reconnects on error so the dashboard self-heals.
func subscribeHub(addr string, reg *domain.HostRegistry, dialOpts []grpc.DialOption) {
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
			id, ms := transport.FromSnapshot(snap)
			hid := domain.HostID(id)
			reg.Enroll(domain.Host{ID: hid, Hostname: id, DisplayName: id}, time.Now())
			reg.Observe(hid, ms, time.Now())
		}
		_ = conn.Close()
		time.Sleep(time.Second)
	}
}
