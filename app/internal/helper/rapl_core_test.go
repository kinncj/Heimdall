// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build linux

package helper

import "testing"

// The RAPL "core" (pp0) subdomain is the one that maps to power.cpu; package and
// uncore/dram must not be mistaken for it.
func TestIsCoreDomain(t *testing.T) {
	if !isCoreDomain(" core\n") {
		t.Error(`"core" should map to power.cpu`)
	}
	for _, n := range []string{"package-0", "uncore", "dram", "psys", ""} {
		if isCoreDomain(n) {
			t.Errorf("%q is not the core domain", n)
		}
	}
}
