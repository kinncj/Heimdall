package pane

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRectContains(t *testing.T) {
	r := Rect{X: 2, Y: 3, W: 4, H: 5}
	if !r.Contains(2, 3) || !r.Contains(5, 7) {
		t.Fatal("corner points should be inside")
	}
	if r.Contains(6, 3) || r.Contains(2, 8) || r.Contains(1, 3) {
		t.Fatal("outside points reported inside")
	}
}

func TestFocusWrapsBothDirections(t *testing.T) {
	g := NewGroup(
		New("a", &staticSrc{rows: makeRows(2)}),
		New("b", &staticSrc{rows: makeRows(2)}),
		New("c", &staticSrc{rows: makeRows(2)}),
	)
	if g.Focus() != 0 || !g.Panes()[0].Focused() {
		t.Fatal("first pane should start focused")
	}
	g.Update(key("tab"), 20)
	g.Update(key("tab"), 20)
	if g.Focus() != 2 {
		t.Fatalf("focus=%d want 2", g.Focus())
	}
	g.Update(key("tab"), 20) // wrap to 0
	if g.Focus() != 0 {
		t.Fatalf("tab did not wrap: focus=%d want 0", g.Focus())
	}
	g.Update(keyShiftTab(), 20) // wrap back to 2
	if g.Focus() != 2 {
		t.Fatalf("shift+tab did not wrap: focus=%d want 2", g.Focus())
	}
	// Exactly one pane focused.
	n := 0
	for _, p := range g.Panes() {
		if p.Focused() {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d panes focused, want exactly 1", n)
	}
}

func keyShiftTab() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyShiftTab} }

// The whole point of the small-screen invariant: focusing a pane below the fold
// page-scrolls the canvas so that pane comes into view.
func TestFocusingPaneBelowFoldPageScrolls(t *testing.T) {
	// Three tall panes; the viewport is far smaller than their sum.
	g := NewGroup(
		New("a", &staticSrc{rows: makeRows(10)}),
		New("b", &staticSrc{rows: makeRows(10)}),
		New("c", &staticSrc{rows: makeRows(10)}),
	)
	const h = 12
	if g.PageScroll() != 0 {
		t.Fatalf("initial pageScroll=%d want 0", g.PageScroll())
	}
	g.Update(key("tab"), h) // focus b
	g.Update(key("tab"), h) // focus c — well below the fold
	if g.PageScroll() == 0 {
		t.Fatal("focusing a pane below the fold should page-scroll the canvas")
	}
	// The focused pane's canvas top must be within the visible page window.
	slots, _ := g.layout(50, h)
	top := slots[g.Focus()].top
	if top < g.PageScroll() || top >= g.PageScroll()+h {
		// The pane's active row (top+2) is what must be visible; assert that.
		active := top + 2
		if active < g.PageScroll() || active >= g.PageScroll()+h {
			t.Fatalf("focused pane active row %d not visible in page window [%d,%d)",
				active, g.PageScroll(), g.PageScroll()+h)
		}
	}
}

func TestMouseWheelScrollsPaneUnderPointer(t *testing.T) {
	m := testMode(t)
	a := New("a", &staticSrc{rows: makeRows(40)})
	b := New("b", &staticSrc{rows: makeRows(40)})
	g := NewGroup(a, b)
	// Render so panes get on-screen boxes.
	g.View(m, Rect{X: 0, Y: 0, W: 50, H: 30})

	// Find a point inside pane a's box and wheel there; a should scroll, b should not.
	box := a.Box()
	if box.H == 0 {
		t.Fatal("pane a has no on-screen box after render")
	}
	before := a.Scroll()
	g.Mouse(wheel(tea.MouseButtonWheelDown, box.X+1, box.Y+1), 30)
	if a.Scroll() == before {
		t.Fatal("wheel over pane a did not scroll it")
	}
	if b.Scroll() != 0 {
		t.Fatal("wheel over pane a wrongly scrolled pane b")
	}
}

func TestMouseClickFocusesPaneUnderPointer(t *testing.T) {
	m := testMode(t)
	a := New("a", &staticSrc{rows: makeRows(6)})
	b := New("b", &staticSrc{rows: makeRows(6)})
	g := NewGroup(a, b)
	g.View(m, Rect{X: 0, Y: 0, W: 50, H: 40})
	if g.Focus() != 0 {
		t.Fatalf("focus=%d want 0", g.Focus())
	}
	box := b.Box()
	if box.H == 0 {
		t.Fatal("pane b off-screen")
	}
	g.Mouse(click(box.X+1, box.Y+1), 40)
	if g.Focus() != 1 {
		t.Fatalf("click on pane b did not focus it: focus=%d", g.Focus())
	}
}

func TestGroupPageAffordanceOnOverflow(t *testing.T) {
	m := testMode(t)
	g := NewGroup(
		New("a", &staticSrc{rows: makeRows(20)}),
		New("b", &staticSrc{rows: makeRows(20)}),
	)
	out := g.View(m, Rect{X: 0, Y: 0, W: 50, H: 12})
	if !strings.Contains(out, "page") {
		t.Fatalf("overflowing group should show a page affordance:\n%s", out)
	}
}

// A grid row lays panes side by side; Tab flows reading-order (row-major) and each
// column gets its own on-screen box so the mouse can tell them apart.
func TestGridReadingOrderAndColumns(t *testing.T) {
	m := testMode(t)
	a := New("a", &staticSrc{rows: makeRows(3)})
	b := New("b", &staticSrc{rows: makeRows(3)})
	c := New("c", &staticSrc{rows: makeRows(3)})
	g := NewGrid([][]*Pane{{a, b}, {c}})

	// Reading order: a, b, c.
	got := g.Panes()
	if len(got) != 3 || got[0] != a || got[1] != b || got[2] != c {
		t.Fatalf("reading order wrong: %v", got)
	}
	g.Update(key("tab"), 40) // a -> b
	if g.Focused() != b {
		t.Fatalf("tab did not move to b")
	}
	g.Update(key("tab"), 40) // b -> c
	if g.Focused() != c {
		t.Fatalf("tab did not move to c")
	}

	g.View(m, Rect{X: 0, Y: 0, W: 60, H: 40})
	if a.Box().W == 0 || b.Box().W == 0 {
		t.Fatal("row panes should have on-screen boxes")
	}
	if b.Box().X <= a.Box().X {
		t.Fatalf("second column should sit to the right: a.X=%d b.X=%d", a.Box().X, b.Box().X)
	}
	// A click in column b focuses b, not a.
	g.Mouse(click(b.Box().X+1, b.Box().Y+1), 40)
	if g.Focused() != b {
		t.Fatalf("click in column b focused %s", g.Focused().Title())
	}
}

// With page focus, the whole view is the first stop: arrows scroll the whole page,
// Tab steps into each pane, and the cycle wraps back to the whole-page stop.
func TestPageFocusFirstStopAndWrap(t *testing.T) {
	g := NewGrid([][]*Pane{
		{New("a", &staticSrc{rows: makeRows(15)})},
		{New("b", &staticSrc{rows: makeRows(15)})},
	}).EnablePageFocus()

	const h = 12
	if !g.WholePageFocused() {
		t.Fatal("page-focus group should start on the whole-page stop")
	}
	if g.Focused() != nil {
		t.Fatal("no pane should be focused on the whole-page stop")
	}
	// Arrows scroll the whole page while nothing is focused.
	g.Update(key("down"), h)
	g.Update(key("down"), h)
	if g.PageScroll() == 0 {
		t.Fatal("arrows should scroll the whole page on the whole-page stop")
	}
	// Tab steps into pane 0, then pane 1, then wraps back to the whole-page stop.
	g.Update(key("tab"), h)
	if g.Focus() != 0 {
		t.Fatalf("first Tab should focus pane 0, got %d", g.Focus())
	}
	g.Update(key("tab"), h)
	if g.Focus() != 1 {
		t.Fatalf("second Tab should focus pane 1, got %d", g.Focus())
	}
	g.Update(key("tab"), h) // wrap back to whole page
	if !g.WholePageFocused() {
		t.Fatalf("Tab should wrap back to the whole-page stop, focus=%d", g.Focus())
	}
	// Shift-Tab from the whole-page stop goes to the last pane.
	g.Update(keyShiftTab(), h)
	if g.Focus() != 1 {
		t.Fatalf("shift+Tab from whole-page should focus the last pane, got %d", g.Focus())
	}
}

func TestPageFocusWheelScrollsWholePage(t *testing.T) {
	m := testMode(t)
	g := NewGrid([][]*Pane{
		{New("a", &staticSrc{rows: makeRows(20)})},
		{New("b", &staticSrc{rows: makeRows(20)})},
	}).EnablePageFocus()
	g.View(m, Rect{X: 0, Y: 0, W: 50, H: 12})
	before := g.PageScroll()
	g.Mouse(wheel(tea.MouseButtonWheelDown, 5, 5), 12)
	if g.PageScroll() <= before {
		t.Fatal("wheel on the whole-page stop should scroll the whole page")
	}
}

func wheel(btn tea.MouseButton, x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: btn, Action: tea.MouseActionPress}
}

func click(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
}
