package pane

import (
	"strings"
	"testing"
)

func TestWindowNoScrollWhenFits(t *testing.T) {
	m := testMode(t)
	lines := makeRows(4)
	out, off := Window(m, lines, 3, -1, 10)
	if off != 0 || len(out) != 4 {
		t.Fatalf("content that fits should not scroll: off=%d len=%d", off, len(out))
	}
	if strings.Contains(strings.Join(out, "\n"), "more") {
		t.Fatal("no affordance expected when everything fits")
	}
}

func TestWindowAffordancesOnOverflow(t *testing.T) {
	m := testMode(t)
	lines := makeRows(20)
	// Middle of a long list: both edges show a "more" affordance.
	out, off := Window(m, lines, 8, -1, 6)
	if off != 8 {
		t.Fatalf("offset=%d want 8", off)
	}
	if !strings.Contains(out[0], "more") || !strings.Contains(out[len(out)-1], "more") {
		t.Fatalf("expected both edges to show affordances:\n%s", strings.Join(out, "\n"))
	}
}

// The selection-visibility invariant that used to live in the dashboard's
// keepSelVisible: whatever the selection, Window keeps it inside the viewport and
// one row clear of the ▲/▼ "more" markers on scrolled edges.
func TestWindowKeepsSelectionVisible(t *testing.T) {
	m := testMode(t)
	const total, height = 40, 6
	lines := makeRows(total)
	for sel := 0; sel < total; sel++ {
		_, off := Window(m, lines, 0, sel, height)
		top, bottom := off, off+height-1
		if sel < top || sel > bottom {
			t.Fatalf("sel %d not in window [%d,%d]", sel, top, bottom)
		}
		if off > 0 && sel == top {
			t.Fatalf("sel %d sits on the top ▲more row (off=%d)", sel, off)
		}
		if off+height < total && sel == bottom {
			t.Fatalf("sel %d sits on the bottom ▼more row (off=%d)", sel, off)
		}
	}
}
