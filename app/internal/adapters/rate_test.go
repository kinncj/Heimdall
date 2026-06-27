// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import "testing"

func TestDeltaRateMB(t *testing.T) {
	// 2,000,000 bytes over 2s = 1 MB/s.
	if got := deltaRateMB(0, 2_000_000, 2); got != 1 {
		t.Fatalf("1 MB/s: got %v", got)
	}
	// A counter reset (cur < prev) must yield 0, not a uint underflow spike.
	if got := deltaRateMB(5_000_000, 100, 1); got != 0 {
		t.Fatalf("counter reset should be 0, got %v", got)
	}
	// A non-positive interval must yield 0.
	if got := deltaRateMB(0, 1_000_000, 0); got != 0 {
		t.Fatalf("zero interval should be 0, got %v", got)
	}
}
