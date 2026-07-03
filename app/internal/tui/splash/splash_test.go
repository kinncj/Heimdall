// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package splash

import "testing"

// Ghostty over SSH forwards TERM but not TERM_PROGRAM, so rasterm alone reports the
// terminal as image-incapable and the splash falls back to ASCII. kittyCapable must
// still recognise it from TERM=xterm-ghostty.
func TestKittyCapableDetectsGhosttyByTERM(t *testing.T) {
	// Clear the identifiers rasterm keys off, to isolate the TERM fallback.
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")

	t.Setenv("TERM", "xterm-ghostty")
	if !kittyCapable() {
		t.Error("TERM=xterm-ghostty (e.g. Ghostty over SSH) should be Kitty-capable")
	}

	// Case-insensitive / whitespace tolerant.
	t.Setenv("TERM", "  XTERM-GHOSTTY ")
	if !kittyCapable() {
		t.Error("TERM match should tolerate case and surrounding whitespace")
	}

	// A plain terminal with none of the identifiers must not be treated as capable.
	t.Setenv("TERM", "xterm-256color")
	if kittyCapable() {
		t.Error("xterm-256color with no image identifiers should not be Kitty-capable")
	}
}

// The existing TERM_PROGRAM path (macOS Ghostty, local) must keep working.
func TestKittyCapableKeepsTermProgramPath(t *testing.T) {
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "ghostty")
	if !kittyCapable() {
		t.Error("TERM_PROGRAM=ghostty (local macOS Ghostty) should stay Kitty-capable")
	}
}
