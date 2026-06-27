// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import "testing"

// Yggdrasil: the origin hub is stamped as an authoritative "hub" label and
// cannot be spoofed by an incoming label; other tags are preserved.
func TestWithOriginIsAuthoritative(t *testing.T) {
	got := withOrigin(map[string]string{"env": "prod", "hub": "spoofed"}, "real-hub")
	if got["hub"] != "real-hub" {
		t.Errorf("hub = %q, want real-hub (origin is authoritative)", got["hub"])
	}
	if got["env"] != "prod" {
		t.Errorf("env label lost: %v", got)
	}
}
