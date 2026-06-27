// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import "heimdall/app/internal/options"

// dashboardCatalog declares the persistent dashboard settings. Action/mode flags
// (--snapshot, --demo, --detail, --control/--run/--tail, --splash) stay separate
// since they are not saved.
func dashboardCatalog() options.Catalog {
	return options.NewCatalog(
		options.Define("hub").Default("localhost:9090").
			Help("hub address to subscribe to; 'auto' discovers via mDNS (Ratatoskr)").
			Ask("Hub address to subscribe to"),
		options.Define("discover").Of(options.KindToggle).Default("false").
			Help("auto-discover the hub via mDNS when --hub is auto/unset (Ratatoskr)"),
		options.Define("discover-seed").
			Help("fallback hub address for discovery on overlay networks (Tailscale, etc.)"),
		options.Define("mode").Default("dark").Help("theme mode: dark|light"),
		options.Define("token").Of(options.KindSecret).Env("HEIMDALL_TOKEN").
			Help("enrollment token presented to the hub (env HEIMDALL_TOKEN)"),
		options.Define("tls").Of(options.KindToggle).Default("false").Help("connect to the hub over TLS"),
		options.Define("tls-ca").Help("PEM CA bundle to trust (default: system roots)"),
		options.Define("tls-server-name").Help("override the server name verified in the hub certificate"),
		options.Define("tls-insecure").Of(options.KindToggle).Default("false").Help("skip hub certificate verification (dev only)"),
		options.Define("purge-after").Of(options.KindSpan).Default("15m").
			Help("drop a host from the view after it has been unseen this long (0 disables)"),
	)
}

// resolveDashboard folds defaults, config, env, and flags, running the first-run
// wizard on a fresh terminal with no config and no flags.
func resolveDashboard(cat options.Catalog) options.Resolved {
	return options.Resolve("dashboard", cat,
		"heimdall-dashboard — first-run setup",
		"The dashboard subscribes to a hub and renders the fleet (--demo needs no hub).")
}
