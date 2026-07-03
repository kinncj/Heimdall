// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import "testing"

// keepSelVisible is the fix for the selection lists (command picker, log-source
// list, sort) whose cursor used to drift out of the scroll window.
func TestKeepSelVisible(t *testing.T) {
	cases := []struct {
		name                     string
		offset, sel, body, total int
		want                     int
	}{
		{"fits: no scroll", 5, 3, 10, 4, 0},
		{"cursor at top stays", 0, 0, 5, 20, 0},
		{"cursor below window scrolls down", 0, 9, 5, 20, 6}, // sel-body+2 = 9-5+2
		{"cursor above window scrolls up", 10, 4, 5, 20, 3},  // sel-1
		{"clamp to last window", 0, 19, 5, 20, 15},           // total-body
		{"clamp not negative", 3, 0, 5, 20, 0},
	}
	for _, c := range cases {
		if got := keepSelVisible(c.offset, c.sel, c.body, c.total); got != c.want {
			t.Errorf("%s: keepSelVisible(%d,%d,%d,%d)=%d want %d",
				c.name, c.offset, c.sel, c.body, c.total, got, c.want)
		}
	}
}

// The invariant that matters: whatever the selection, it lands inside the visible
// window, one row clear of the ↑/↓ "more" indicators on scrolled edges.
func TestKeepSelVisibleKeepsCursorInWindow(t *testing.T) {
	const body, total = 6, 40
	for sel := 0; sel < total; sel++ {
		off := keepSelVisible(0, sel, body, total)
		top, bottom := off, off+body-1
		// Selected row is within the window.
		if sel < top || sel > bottom {
			t.Fatalf("sel %d not in window [%d,%d]", sel, top, bottom)
		}
		// And not sitting on an indicator row (unless at the list's true edge).
		if off > 0 && sel == top {
			t.Fatalf("sel %d sits on the top ↑more indicator (off=%d)", sel, off)
		}
		if off+body < total && sel == bottom {
			t.Fatalf("sel %d sits on the bottom ↓more indicator (off=%d)", sel, off)
		}
	}
}
