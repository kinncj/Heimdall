package pane

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/tui/theme"
)

func testMode(t *testing.T) theme.Mode {
	t.Helper()
	th, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	m, ok := th.Mode("")
	if !ok {
		t.Fatal("default mode missing")
	}
	return m
}

// --- fake sources ------------------------------------------------------------

// staticSrc is a scroll-only source (implements only Source).
type staticSrc struct{ rows []string }

func (s *staticSrc) Rows() []string { return s.rows }

// listSrc is a selectable, filterable, sortable source.
type listSrc struct {
	all      []string
	query    string
	sortKey  string
	sortKeys []string
}

func (l *listSrc) Rows() []string {
	if l.query == "" {
		return l.all
	}
	var out []string
	for _, r := range l.all {
		if strings.Contains(r, l.query) {
			out = append(out, r)
		}
	}
	return out
}
func (l *listSrc) RowCount() int      { return len(l.Rows()) }
func (l *listSrc) SetFilter(q string) { l.query = q }
func (l *listSrc) SortKeys() []string { return l.sortKeys }
func (l *listSrc) SetSort(k string)   { l.sortKey = k }

func key(s string) tea.KeyMsg {
	switch s {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "/":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	case "s":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// --- pure math ---------------------------------------------------------------

func TestClampOffset(t *testing.T) {
	cases := []struct{ off, total, h, want int }{
		{0, 3, 10, 0},   // fits: no scroll
		{5, 3, 10, 0},   // fits but over-scrolled: clamp to 0
		{99, 20, 5, 15}, // max is total-h
		{-4, 20, 5, 0},  // negative clamps to 0
	}
	for _, c := range cases {
		if got := clampOffset(c.off, c.total, c.h); got != c.want {
			t.Errorf("clampOffset(%d,%d,%d)=%d want %d", c.off, c.total, c.h, got, c.want)
		}
	}
}

func TestKeepVisible(t *testing.T) {
	cases := []struct{ off, target, h, want int }{
		{0, 0, 5, 0}, // already visible
		{0, 4, 5, 0}, // last visible row, no move
		{0, 5, 5, 1}, // one past window: scroll down by 1
		{0, 9, 5, 5}, // far below: window ends at target
		{5, 2, 5, 2}, // above window: window starts at target
	}
	for _, c := range cases {
		if got := keepVisible(c.off, c.target, c.h); got != c.want {
			t.Errorf("keepVisible(%d,%d,%d)=%d want %d", c.off, c.target, c.h, got, c.want)
		}
	}
}

// --- selection follows scroll ------------------------------------------------

func TestSelectionFollowsCursorDown(t *testing.T) {
	src := &listSrc{all: makeRows(20)}
	p := New("L", src)
	p.SetFocused(true)
	const h = 5
	for i := 0; i < 9; i++ { // move cursor to row 9
		p.Update(key("down"), h)
	}
	if p.Selection() != 9 {
		t.Fatalf("selection=%d want 9", p.Selection())
	}
	// Cursor must be inside the window [scroll, scroll+h).
	if p.Scroll() > 9 || 9 >= p.Scroll()+h {
		t.Fatalf("cursor 9 not visible: scroll=%d h=%d", p.Scroll(), h)
	}
}

func TestEndGoesToLastAndKeepsVisible(t *testing.T) {
	src := &listSrc{all: makeRows(20)}
	p := New("L", src)
	p.SetFocused(true)
	p.Update(key("end"), 5)
	if p.Selection() != 19 {
		t.Fatalf("selection=%d want 19", p.Selection())
	}
	if p.Scroll() != 15 {
		t.Fatalf("scroll=%d want 15 (last window)", p.Scroll())
	}
}

func TestScrollOnlyPaneHasNoSelection(t *testing.T) {
	p := New("S", &staticSrc{rows: makeRows(20)})
	if p.Selection() != -1 {
		t.Fatalf("static pane selection=%d want -1", p.Selection())
	}
	p.Update(key("down"), 5)
	if p.Scroll() != 1 {
		t.Fatalf("scroll=%d want 1", p.Scroll())
	}
}

// --- filter ------------------------------------------------------------------

func TestSlashFilterNarrowsAndResets(t *testing.T) {
	src := &listSrc{all: []string{"alpha", "beta", "gamma", "gamble"}}
	p := New("L", src)
	p.SetFocused(true)
	if !p.Update(key("/"), 10) || !p.Filtering() {
		t.Fatal("slash did not open filter")
	}
	for _, r := range "gam" {
		p.Update(key(string(r)), 10)
	}
	rows := src.Rows()
	if len(rows) != 2 {
		t.Fatalf("filtered rows=%d want 2 (%v)", len(rows), rows)
	}
	if p.Selection() != 0 {
		t.Fatalf("selection after filter=%d want 0", p.Selection())
	}
	p.Update(key("esc"), 10)
	if p.Filtering() || src.query != "" {
		t.Fatal("esc did not clear filter")
	}
}

func TestStaticPaneRejectsFilterAndSort(t *testing.T) {
	p := New("S", &staticSrc{rows: makeRows(5)})
	if p.CanFilter() || p.CanSort() {
		t.Fatal("static pane must not advertise filter/sort")
	}
	if p.Update(key("/"), 10) {
		t.Fatal("static pane consumed / (should not filter)")
	}
	if p.Update(key("s"), 10) {
		t.Fatal("static pane consumed s (should not sort)")
	}
}

// --- sort --------------------------------------------------------------------

func TestSortCycles(t *testing.T) {
	src := &listSrc{all: makeRows(3), sortKeys: []string{"cpu", "mem", "pid"}}
	p := New("P", src)
	if p.SortKey() != "cpu" {
		t.Fatalf("initial sort=%q want cpu", p.SortKey())
	}
	p.Update(key("s"), 10)
	if src.sortKey != "mem" || p.SortKey() != "mem" {
		t.Fatalf("after s: src=%q pane=%q want mem", src.sortKey, p.SortKey())
	}
	p.Update(key("s"), 10)
	p.Update(key("s"), 10) // wraps back to cpu
	if p.SortKey() != "cpu" {
		t.Fatalf("sort wrapped to %q want cpu", p.SortKey())
	}
}

// --- affordances / rendering -------------------------------------------------

func TestAffordanceAppearsOnlyWhenOverflowing(t *testing.T) {
	m := testMode(t)
	small := New("S", &staticSrc{rows: makeRows(3)})
	if out := small.View(m, 40, 10); strings.Contains(out, "more") {
		t.Fatalf("no affordance expected when content fits:\n%s", out)
	}
	big := New("B", &staticSrc{rows: makeRows(50)})
	if out := big.View(m, 40, 10); !strings.Contains(out, "more") {
		t.Fatalf("affordance expected when content overflows:\n%s", out)
	}
}

func TestFocusedCursorRowGetsCaret(t *testing.T) {
	m := testMode(t)
	p := New("L", &listSrc{all: makeRows(5)})
	p.SetFocused(true)
	out := p.View(m, 40, 10)
	if !strings.Contains(out, "▸") {
		t.Fatalf("focused selectable pane should render a ▸ caret:\n%s", out)
	}
	p.SetFocused(false)
	if out := p.View(m, 40, 10); strings.Contains(out, "▸") {
		t.Fatalf("unfocused pane must not render a caret:\n%s", out)
	}
}

func makeRows(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "  row " + string(rune('A'+i%26))
	}
	return out
}
