// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import "strings"

// ParseTags parses a "k=v,k2=v2" tag spec into a label map (Realms). Blank
// entries and entries without '=' are skipped; an empty spec yields nil.
func ParseTags(spec string) map[string]string {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}
	out := map[string]string{}
	for _, pair := range strings.Split(spec, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if k = strings.TrimSpace(k); !ok || k == "" {
			continue
		}
		out[k] = strings.TrimSpace(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
