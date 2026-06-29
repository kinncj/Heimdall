// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import "heimdall/app/internal/options"

// daemonCatalog declares every persistent daemon setting once. Each option drives
// a flag, a JSON config field, an optional env var, and (where it has a question)
// a wizard prompt.
func daemonCatalog() options.Catalog {
	return options.NewCatalog(
		options.Define("hub").Default("localhost:9090").
			Help("hub address to stream to; empty or --print prints locally").
			Ask("Hub address to stream to (blank = print locally)"),
		options.Define("name").
			Help("host display name (default: system hostname)").
			Ask("Host display name (blank = system hostname)"),
		options.Define("interval").Of(options.KindSpan).Default("2s").
			Help("sample interval").Ask("Sample interval"),
		options.Define("ping-target").Default("1.1.1.1").
			Help("internet host pinged for reachability/latency").
			Ask("Internet reachability ping target"),
		options.Define("tags").
			Help("host tags as k=v,k2=v2 (Realms), e.g. env=prod,role=db").
			Ask("Host tags (k=v,k2=v2; blank = none)"),
		options.Define("discover").Of(options.KindToggle).Default("false").
			Help("auto-discover the hub via mDNS when --hub is unset (Ratatoskr); --hub auto forces it"),
		options.Define("discover-seed").
			Help("fallback hub address for discovery on overlay networks (Tailscale, etc.) with no multicast"),
		options.Define("json").Of(options.KindToggle).Default("false").
			Help("emit one JSON object per metric (print mode)"),
		options.Define("token").Of(options.KindSecret).Env("HEIMDALL_TOKEN").
			Help("enrollment token presented to the hub (env HEIMDALL_TOKEN)"),
		options.Define("tls").Of(options.KindToggle).Default("false").
			Help("connect to the hub over TLS"),
		options.Define("tls-ca").Help("PEM CA bundle to trust (default: system roots)"),
		options.Define("tls-server-name").Help("override the server name verified in the hub certificate"),
		options.Define("tls-insecure").Of(options.KindToggle).Default("false").
			Help("skip hub certificate verification (dev only)"),
		options.Define("log-source").
			Help("opt-in log sources alias=path,alias2=path2 to tail and push to the hub (Heimdallr's sight)"),
		options.Define("process-interval").Of(options.KindSpan).Default("0s").
			Help("push a process table to the hub at this interval (Heimdallr's sight); 0 = off"),
		options.Define("log-file").Help("operational log destination: unset = stderr; 'false' = disabled; a path = JSON"),
	)
}

// resolveDaemon folds defaults, the saved config, the environment, and flags into
// the effective settings, running the first-run wizard on a fresh terminal.
func resolveDaemon(cat options.Catalog) options.Resolved {
	return options.Resolve("daemon", cat,
		"heimdall-daemon — first-run setup",
		"This host samples its own metrics and streams them to a hub.",
		"Press Enter to accept each [default]; advanced options live in the saved file.")
}
