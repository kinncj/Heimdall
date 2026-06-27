// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package observe (Mímir) renders the hub's live host views as Prometheus/
// OpenMetrics text and keeps a bounded in-memory history of recent samples, so
// existing observability stacks can scrape Heimdall and short-range trends are
// queryable. See ADR 0011.
package observe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"heimdall/app/internal/domain"
)

// RenderOpenMetrics renders host views as Prometheus text-exposition format: a
// gauge family per metric, plus a heimdall_host_up liveness series. Every series
// carries the host id and the host's labels (tags); per-core metrics fan out to
// one series per core. Non-OK and info metrics are not emitted as numeric series.
func RenderOpenMetrics(views []domain.HostView) string {
	families := map[string][]string{}
	add := func(family string, labels map[string]string, value float64) {
		name := "heimdall_" + family
		families[name] = append(families[name],
			name+formatLabels(labels)+" "+strconv.FormatFloat(value, 'g', -1, 64))
	}
	for _, v := range views {
		base := map[string]string{"host": string(v.Host.ID)}
		for k, val := range v.Host.Context.Labels {
			base[k] = val
		}
		up := withLabel(base, "state", v.State.String())
		add("host_up", up, 1)

		for _, m := range v.LastSnapshot {
			if m.Status != domain.StatusOK {
				continue
			}
			family := sanitize(m.Name)
			if m.Kind == domain.KindPerCore {
				for i, c := range m.PerCore {
					add(family, withLabel(base, "core", strconv.Itoa(i)), c)
				}
				continue
			}
			if m.Unit == "info" {
				continue
			}
			add(family, base, m.Gauge)
		}
	}

	names := make([]string, 0, len(families))
	for name := range families {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		fmt.Fprintf(&b, "# TYPE %s gauge\n", name)
		lines := families[name]
		sort.Strings(lines)
		for _, ln := range lines {
			b.WriteString(ln)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// sanitize maps a metric name to a valid Prometheus identifier (cpu.util ->
// cpu_util): anything outside [a-zA-Z0-9_] becomes an underscore.
func sanitize(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// withLabel returns a copy of base with one extra label set.
func withLabel(base map[string]string, key, value string) map[string]string {
	out := make(map[string]string, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	out[key] = value
	return out
}

// formatLabels renders a label set as {k="v",...} with keys sorted and values
// escaped per the exposition format. An empty set renders as "".
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+`="`+escapeLabel(labels[k])+`"`)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func escapeLabel(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	return v
}
