// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"sort"
	"strings"

	"heimdall/app/internal/domain"
)

// groupDim is one grouping axis (Yggdrasil): a name and a key extractor over a
// host view. Adding an axis means registering another dim, not editing a switch.
type groupDim struct {
	name string
	of   func(domain.HostView) string
}

// dimensions returns the grouping axes for the current fleet: origin hub, OS,
// then each distinct tag key. `groupBy` indexes this list with 1..N; 0 = none.
func (m Model) dimensions(hosts []domain.HostView) []groupDim {
	dims := []groupDim{
		{"hub", func(h domain.HostView) string { return labelOr(h, "hub", "(no hub)") }},
		{"os", osOf},
	}
	seen := map[string]bool{"hub": true}
	var keys []string
	for _, h := range hosts {
		for k := range h.Host.Context.Labels {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		k := k
		dims = append(dims, groupDim{k, func(h domain.HostView) string { return labelOr(h, k, "(no "+k+")") }})
	}
	return dims
}

// osOf derives a host's OS family for grouping. The dashboard receives OS only
// as the host.os inventory metric (e.g. "darwin 27.0") — Context.OS is empty
// over the wire — so read the metric first (family = first token) and fall back
// to Context.OS (set in --demo) before "(unknown)".
func osOf(h domain.HostView) string {
	for _, m := range h.LastSnapshot {
		if m.Name == "host.os" && m.Detail != "" {
			if f := strings.Fields(m.Detail); len(f) > 0 {
				return f[0]
			}
		}
	}
	if h.Host.Context.OS != "" {
		return h.Host.Context.OS
	}
	return "(unknown)"
}

func labelOr(h domain.HostView, key, fallback string) string {
	if v := h.Host.Context.Labels[key]; v != "" {
		return v
	}
	return fallback
}

// activeDim returns the current grouping dimension, or false when grouping is off.
func (m Model) activeDim(hosts []domain.HostView) (groupDim, bool) {
	if m.groupBy <= 0 {
		return groupDim{}, false
	}
	dims := m.dimensions(hosts)
	if m.groupBy-1 >= len(dims) {
		return groupDim{}, false
	}
	return dims[m.groupBy-1], true
}

// fieldMatcher is one searchable axis — an adapter over a HostView, mirroring
// groupDim. Adding a scope means registering another matcher, not editing a
// switch (Open/Closed); a host satisfies a scoped term via the matcher whose
// name equals the scope, and a bare term via any matcher.
type fieldMatcher struct {
	name string
	of   func(domain.HostView) string
}

// matchers returns the searchable fields for the current fleet: host name/id,
// hub, os, state, the active grouping dimension (as "group"), then each tag key.
// Tag keys are derived from the fleet like dimensions() — no field is hard-coded
// beyond the four structural ones.
func (m Model) matchers() []fieldMatcher {
	fields := []fieldMatcher{
		{"host", func(h domain.HostView) string { return h.Host.DisplayName + " " + string(h.Host.ID) }},
		{"hub", func(h domain.HostView) string { return labelOr(h, "hub", "") }},
		{"os", osOf},
		{"state", func(h domain.HostView) string { return stateName(h.State) }},
	}
	// "group" is an alias that delegates to the active grouping dimension, reusing
	// the grouping strategy instead of duplicating it. Absent when grouping is off.
	if dim, ok := m.activeDim(m.reg.Hosts()); ok {
		fields = append(fields, fieldMatcher{"group", dim.of})
	}
	seen := map[string]bool{"host": true, "hub": true, "os": true, "state": true, "group": true}
	var keys []string
	for _, h := range m.reg.Hosts() {
		for k := range h.Host.Context.Labels {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		k := k
		fields = append(fields, fieldMatcher{k, func(h domain.HostView) string { return h.Host.Context.Labels[k] }})
	}
	return fields
}

// searchTerm is one parsed filter token: an optional field scope and a value.
type searchTerm struct {
	field string // "" = unscoped (matches any field)
	value string
}

// parseTerms splits the filter into space-separated terms. "field=value" is a
// scoped term only when field names a known matcher; otherwise the whole token
// (any '=' included) is an unscoped value matched against every field.
func parseTerms(filter string, fields map[string]bool) []searchTerm {
	var terms []searchTerm
	for _, tok := range strings.Fields(strings.ToLower(filter)) {
		if i := strings.IndexByte(tok, '='); i > 0 && fields[tok[:i]] {
			terms = append(terms, searchTerm{field: tok[:i], value: tok[i+1:]})
			continue
		}
		terms = append(terms, searchTerm{value: tok})
	}
	return terms
}

// termMatches reports whether a host satisfies one term: a scoped term checks its
// one field; a bare term checks every field. Substring, case-insensitive.
func termMatches(t searchTerm, h domain.HostView, fields []fieldMatcher) bool {
	for _, f := range fields {
		if t.field != "" && f.name != t.field {
			continue
		}
		if strings.Contains(strings.ToLower(f.of(h)), t.value) {
			return true
		}
	}
	return false
}

// matchesFilter reports whether a host satisfies every parsed term (AND). An
// empty filter yields no terms and matches everything.
func (m Model) matchesFilter(h domain.HostView, fields []fieldMatcher, terms []searchTerm) bool {
	for _, t := range terms {
		if !termMatches(t, h, fields) {
			return false
		}
	}
	return true
}

// orderedHosts returns the hosts in display order: filtered, and when grouped,
// sorted so each group is contiguous. The parallel []string holds each host's
// group label (nil when not grouped) for section headers.
func (m Model) orderedHosts() ([]domain.HostView, []string) {
	fields := m.matchers()
	names := make(map[string]bool, len(fields))
	for _, f := range fields {
		names[f.name] = true
	}
	terms := parseTerms(m.filter, names)

	var hosts []domain.HostView
	for _, h := range m.reg.Hosts() {
		if m.matchesFilter(h, fields, terms) {
			hosts = append(hosts, h)
		}
	}
	dim, grouped := m.activeDim(m.reg.Hosts())
	if !grouped {
		return hosts, nil
	}
	sort.SliceStable(hosts, func(i, j int) bool {
		if ki, kj := dim.of(hosts[i]), dim.of(hosts[j]); ki != kj {
			return ki < kj
		}
		return hosts[i].Host.ID < hosts[j].Host.ID
	})
	labels := make([]string, len(hosts))
	for i, h := range hosts {
		labels[i] = dim.of(h)
	}
	return hosts, labels
}

// orderedList is orderedHosts without the group labels.
func (m Model) orderedList() []domain.HostView {
	hosts, _ := m.orderedHosts()
	return hosts
}

// alertCount returns how many of the given hosts have at least one firing alert.
func alertCount(hosts []domain.HostView) int {
	n := 0
	for _, h := range hosts {
		if len(h.Alerts) > 0 {
			n++
		}
	}
	return n
}
