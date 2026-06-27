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
)

// version is the Heimdall build version, set via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	sock := flag.String("socket", helper.DefaultSocketPath(), "unix socket to listen on")
	demo := flag.Bool("demo", false, "serve canned sample metrics (no root needed; for trying the needs-helper UI)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("heimdall-helper", version)
		return
	}

	collect := helper.PrivilegedMetrics
	if *demo {
		collect = demoMetrics
	}

	srv := &helper.Server{SockPath: *sock, Collect: collect}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Fprintln(os.Stderr, "heimdall-helper: shutting down")
		cancel()
	}()

	fmt.Fprintf(os.Stderr, "heimdall-helper: serving privileged metrics on %s (demo=%t)\n", *sock, *demo)
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
