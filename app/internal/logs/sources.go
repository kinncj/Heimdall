// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package logs is the opt-in, rate-limited log tailing plane. Only explicitly
// registered source aliases can be tailed; with no sources configured, log
// streaming stays off. It runs on its own gRPC service (LogStreamService),
// independent of the metric stream, so log volume can never starve metrics.
package logs

import (
	"fmt"
	"sort"
	"strings"
)

// Sources is the opt-in registry of tail-able log sources: alias -> file path.
// An unknown alias yields nothing; an empty registry means log streaming is off.
type Sources map[string]string

// Resolve returns the file path registered for an alias.
func (s Sources) Resolve(alias string) (string, bool) {
	p, ok := s[alias]
	return p, ok
}

// Aliases lists the registered aliases, sorted.
func (s Sources) Aliases() []string {
	out := make([]string, 0, len(s))
	for a := range s {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

// ParseSources parses "alias=path,alias2=path2" into a registry. Blank input
// yields an empty (off) registry.
func ParseSources(spec string) (Sources, error) {
	out := Sources{}
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return out, nil
	}
	for _, pair := range strings.Split(spec, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" || strings.TrimSpace(kv[1]) == "" {
			return nil, fmt.Errorf("invalid log source %q (want alias=path)", pair)
		}
		out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return out, nil
}
