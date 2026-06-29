// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Command heimdall-helper is the optional privileged metrics unit. It runs as
// its own privileged process and serves power, GPU, and thermal readings to the
// unprivileged daemon over a local unix socket, so the daemon never runs as root.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/helper"
	"heimdall/app/internal/options"
	"heimdall/app/internal/selfupdate"
)

// version is the Heimdall build version, set via -ldflags "-X main.version=…".
var version = "dev"

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, `heimdall-helper — Heimdall privileged sidecar (metrics + commands)

Runs as a privileged process (sudo) and serves power, GPU, and thermal readings
to the local unprivileged daemon over a unix socket, so the daemon never runs as
root. It also runs the privileged, allow-listed commands the daemon delegates
(e.g. dmesg, journal.tail) for the v2 on-demand control plane — enforcing its OWN
allow-list, never trusting the daemon. Install it on a host only where you want
power/GPU/temperature or those privileged diagnostics.

Usage:
  sudo heimdall-helper [flags]
  heimdall-helper update          update to the latest release
  heimdall-helper --version       print version

First run with no config and no flags starts a short setup wizard and saves the
result; any flag you pass is saved to the config file too.

Flags:
`)
	flag.PrintDefaults()
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := selfupdate.Run("helper", version); err != nil {
			fmt.Fprintln(os.Stderr, "heimdall-helper:", err)
			os.Exit(1)
		}
		return
	}
	showVersion := flag.Bool("version", false, "print version and exit")
	demo := flag.Bool("demo", false, "serve canned sample metrics (no root needed; for trying the needs-helper UI)")
	cat := options.NewCatalog(
		options.Define("socket").Default(helper.DefaultSocketPath()).
			Help("unix socket to listen on").Ask("Unix socket path"),
	)
	cat.Register(flag.CommandLine)
	flag.Usage = usage
	flag.Parse()
	if *showVersion {
		fmt.Println("heimdall-helper", version)
		return
	}

	cfg := options.Resolve("helper", cat,
		"heimdall-helper — first-run setup",
		"Serves privileged power/GPU/thermal metrics to the local daemon over a unix socket.")
	sock := cfg.Text("socket")

	collect := helper.PrivilegedMetrics
	if *demo {
		collect = demoMetrics
	}

	srv := &helper.Server{SockPath: sock, Collect: collect}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Fprintln(os.Stderr, "heimdall-helper: shutting down")
		cancel()
	}()

	fmt.Fprintf(os.Stderr, "heimdall-helper: serving privileged metrics on %s (demo=%t)\n", sock, *demo)
	if err := srv.Serve(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "heimdall-helper:", err)
		os.Exit(1)
	}
}

func demoMetrics(context.Context) []domain.Metric {
	return []domain.Metric{
		{Name: "power.cpu", Unit: "watts", Status: domain.StatusOK, Gauge: 8.4},
		{Name: "power.gpu", Unit: "watts", Status: domain.StatusOK, Gauge: 3.1},
		{Name: "power.pkg", Unit: "watts", Status: domain.StatusOK, Gauge: 12.7},
		{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: 41},
		{Name: "gpu.vram", Unit: "percent", Status: domain.StatusOK, Gauge: 36},
		{Name: "gpu.temp", Unit: "celsius", Status: domain.StatusOK, Gauge: 54},
	}
}
