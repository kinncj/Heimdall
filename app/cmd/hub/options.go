// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"flag"
	"fmt"
	"strings"

	"heimdall/app/internal/options"
)

// hubCatalog declares every persistent hub setting once.
func hubCatalog() options.Catalog {
	return options.NewCatalog(
		options.Define("listen").Default("127.0.0.1:9090").
			Help("gRPC listen address(es), comma-separated to bind several interfaces").
			Ask("Listen address(es) (comma-separated; 0.0.0.0:9090 = all interfaces)"),
		options.Define("token").Of(options.KindSecret).Env("HEIMDALL_TOKEN").
			Help("required enrollment token (env HEIMDALL_TOKEN); empty disables auth").
			Ask("Enrollment token required of daemons/dashboards (blank = no auth)"),
		options.Define("tls-cert").Help("PEM server certificate; enables TLS with tls-key"),
		options.Define("tls-key").Help("PEM server private key; enables TLS with tls-cert"),
		options.Define("id").Default(defaultHubID()).Help("this hub's federation id"),
		options.Define("stale-after").Of(options.KindSpan).Default("10s").Help("mark a host stale after no updates for this long"),
		options.Define("offline-after").Of(options.KindSpan).Default("30s").Help("mark a host offline after no updates for this long"),
		options.Define("purge-after").Of(options.KindSpan).Default("15m").Help("drop a host after it has been unseen this long (0 disables)"),
		options.Define("upstream").Help("parent hub address to relay this hub's hosts to (federation)"),
		options.Define("upstream-token").Of(options.KindSecret).Env("HEIMDALL_UPSTREAM_TOKEN").Help("enrollment token for the upstream hub"),
		options.Define("upstream-tls").Of(options.KindToggle).Default("false").Help("relay to the upstream hub over TLS"),
		options.Define("upstream-tls-ca").Help("PEM CA bundle to trust for the upstream hub"),
		options.Define("upstream-tls-server-name").Help("override the server name verified in the upstream certificate"),
		options.Define("upstream-tls-insecure").Of(options.KindToggle).Default("false").Help("skip upstream certificate verification (dev only)"),
		options.Define("relay-interval").Of(options.KindSpan).Default("2s").Help("how often to relay hosts upstream"),
	)
}

// listenAddrs splits a comma-separated --listen value into bind addresses,
// falling back to localhost.
func listenAddrs(s string) []string {
	out := make([]string, 0, 2)
	for _, a := range strings.Split(s, ",") {
		if a = strings.TrimSpace(a); a != "" {
			out = append(out, a)
		}
	}
	if len(out) == 0 {
		return []string{"127.0.0.1:9090"}
	}
	return out
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, `heimdall-hub — Heimdall central server (metric ingest + dashboard fan-out)

Daemons stream metrics to the hub; dashboards subscribe to it for live updates.
Run one hub on your monitoring station; point daemons and dashboards at it.

Usage:
  heimdall-hub [flags]
  heimdall-hub update             update to the latest release
  heimdall-hub --version          print version

By default the hub listens on 127.0.0.1:9090 (local only). To accept daemons from
other machines, set --listen to a reachable address (e.g. 0.0.0.0:9090) — ideally
with --token. --listen takes a comma-separated list to bind several interfaces.

First run with no config and no flags starts a setup wizard and saves the result;
any flag you pass is saved too.

Flags:
`)
	flag.PrintDefaults()
}
