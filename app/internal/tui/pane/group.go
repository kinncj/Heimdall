package pane

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/tui/theme"
)

// Group stacks panes vertically and adds Tab focus, a focus ring, two-level
// page+pane scroll, and pointer-targeted mouse routing.
//
// Two-level scroll: each pane scrolls its own overflow inside its box; the Group
// page-scrolls the whole stack when it is taller than the viewport. Both use the
// same keepVisible rule, so the focused pane's active row is never hidden.
type Group struct {
	panes      []*Pane
	focus      int
	pageScroll int
	box        Rect
	lastHeight int
}

// NewGroup builds a group over panes; the first is focused.
func NewGroup(panes ...*Pane) *Group {
	g := &Group{panes: panes}
	g.applyFocus()
	return g
}

// Panes returns the group's panes in order.
func (g *Group) Panes() []*Pane { return g.panes }

// Focus is the index of the focused pane.
func (g *Group) Focus() int { return g.focus }

// Focused returns the focused pane, or nil if the group is empty.
func (g *Group) Focused() *Pane {
	if len(g.panes) == 0 {
		return nil
	}
	return g.panes[g.focus]
}

// PageScroll is the current page (canvas) scroll offset in rows.
func (g *Group) PageScroll() int { return g.pageScroll }

func (g *Group) applyFocus() {
	for i, p := range g.panes {
		p.SetFocused(i == g.focus)
	}
}

// FocusNext / FocusPrev cycle focus with wrap-around and page-scroll the canvas so
// the newly focused pane's active row comes into view.
func (g *Group) FocusNext() { g.setFocus(g.focus + 1) }
func (g *Group) FocusPrev() { g.setFocus(g.focus - 1) }

func (g *Group) setFocus(i int) {
	n := len(g.panes)
	if n == 0 {
		return
	}
	g.focus = ((i % n) + n) % n // wrap both directions
	g.applyFocus()
	g.ensureFocusedVisible(g.lastHeight)
}

// Update routes a key: Tab/Shift-Tab move focus, everything else goes to the
// focused pane. Reports whether the key was consumed.
func (g *Group) Update(msg tea.KeyMsg, height int) bool {
	g.lastHeight = height
	switch msg.String() {
	case "tab":
		g.FocusNext()
		return true
	case "shift+tab":
		g.FocusPrev()
		return true
	}
	fp := g.Focused()
	if fp == nil {
		return false
	}
	inner := g.innerHeight(height)
	handled := fp.Update(msg, inner)
	if handled {
		g.ensureFocusedVisible(height)
	}
	return handled
}

// Mouse routes a wheel/click by pointer position: the wheel scrolls the pane under
// the pointer, a click focuses it. Reports whether the event was consumed.
func (g *Group) Mouse(msg tea.MouseMsg, height int) bool {
	g.lastHeight = height
	idx := g.paneAt(msg.X, msg.Y)
	if idx < 0 {
		return false
	}
	inner := g.innerHeight(height)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		g.panes[idx].Wheel(-1, inner)
		if idx == g.focus {
			g.ensureFocusedVisible(height)
		}
		return true
	case tea.MouseButtonWheelDown:
		g.panes[idx].Wheel(1, inner)
		if idx == g.focus {
			g.ensureFocusedVisible(height)
		}
		return true
	}
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
		g.focus = idx
		g.applyFocus()
		g.ensureFocusedVisible(height)
		return true
	}
	return false
}

// paneAt returns the index of the pane whose last-rendered box contains (x,y), or
// -1. Panes scrolled off the top/bottom of the page have boxes that do not match.
func (g *Group) paneAt(x, y int) int {
	for i, p := range g.panes {
		if p.Box().Contains(x, y) {
			return i
		}
	}
	return -1
}

// --- layout ------------------------------------------------------------------

const paneChrome = 3 // top border + title row + bottom border

// slot is one pane's placement on the virtual canvas.
type slot struct {
	p     *Pane
	top   int // canvas row of the pane's top border
	inner int // content rows allocated to the pane
	h     int // total pane height (inner + chrome)
}

// innerHeight is the largest content area a single pane may occupy — one viewport
// minus its own chrome. A pane taller than this scrolls internally.
func (g *Group) innerHeight(height int) int {
	h := height - paneChrome
	if h < 1 {
		h = 1
	}
	return h
}

// layout stacks the panes on the canvas, capping each pane's content at one
// viewport so a single tall pane scrolls internally rather than swallowing the page.
func (g *Group) layout(height int) ([]slot, int) {
	maxInner := g.innerHeight(height)
	slots := make([]slot, 0, len(g.panes))
	top := 0
	for _, p := range g.panes {
		natural := len(p.src.Rows())
		if natural < 1 {
			natural = 1
		}
		inner := natural
		if inner > maxInner {
			inner = maxInner
		}
		h := inner + paneChrome
		slots = append(slots, slot{p: p, top: top, inner: inner, h: h})
		top += h
	}
	return slots, top // top == total canvas height
}

// ensureFocusedVisible page-scrolls the canvas so the focused pane's active row is
// on screen. Same keepVisible rule the panes use for their own cursors.
func (g *Group) ensureFocusedVisible(height int) {
	if height < 1 || len(g.panes) == 0 {
		return
	}
	slots, total := g.layout(height)
	s := slots[g.focus]
	// Row of the focused pane's active row, in canvas coordinates: pane top +
	// border + title, then the active row's position within the pane's own window.
	within := s.p.ActiveRow() - s.p.Scroll()
	if within < 0 {
		within = 0
	}
	if within > s.inner-1 {
		within = s.inner - 1
	}
	canvasRow := s.top + 2 + within
	g.pageScroll = clampOffset(keepVisible(g.pageScroll, canvasRow, height), total, height)
}

// --- rendering ---------------------------------------------------------------

// View renders the stacked panes into a height-row window at the given origin,
// page-scrolled so the focused pane's active row is visible, with ▲/▼ page
// affordances on the edges when the canvas overflows. box origin (x,y) lets the
// mouse map screen coordinates back to panes.
func (g *Group) View(m theme.Mode, box Rect) string {
	g.box = box
	g.lastHeight = box.H
	if len(g.panes) == 0 {
		return ""
	}
	slots, total := g.layout(box.H)
	g.ensureFocusedVisible(box.H)
	g.pageScroll = clampOffset(g.pageScroll, total, box.H)

	// Render every pane frame to canvas lines, tagging each line's owning pane so
	// we can set on-screen boxes after page-scrolling.
	type tagged struct {
		line string
		pane int
	}
	canvas := make([]tagged, 0, total)
	for i, s := range slots {
		frame := s.p.View(m, box.W-2, s.inner) // -2 for the border columns
		body := s.p.Frame(m, frame, box.W-2)
		for _, ln := range strings.Split(body, "\n") {
			canvas = append(canvas, tagged{line: ln, pane: i})
		}
	}
	// Normalise: the canvas may be shorter/longer than `total` by rounding; window
	// against its real length.
	realTotal := len(canvas)
	offset := clampOffset(g.pageScroll, realTotal, box.H)
	g.pageScroll = offset

	end := offset + box.H
	if end > realTotal {
		end = realTotal
	}
	view := canvas[offset:end]

	// Record each pane's on-screen box for mouse hit-testing.
	g.setBoxes(slots, offset, box)

	out := make([]string, len(view))
	for i, t := range view {
		out[i] = t.line
	}
	if offset > 0 {
		out = append([]string{g.pageAffordance(m, offset+1, realTotal, true)}, out...)
		if len(out) > box.H {
			out = out[:box.H]
		}
	}
	if end < realTotal {
		out = append(out, g.pageAffordance(m, end, realTotal, false))
		if len(out) > box.H {
			out = out[len(out)-box.H:]
		}
	}
	return strings.Join(out, "\n")
}

// setBoxes assigns each pane its on-screen rectangle given the page offset, so the
// mouse can target it. Panes fully scrolled out get a zero box (no hits).
func (g *Group) setBoxes(slots []slot, offset int, box Rect) {
	for _, s := range slots {
		screenTop := box.Y + (s.top - offset)
		visTop := screenTop
		visBottom := screenTop + s.h
		if visBottom <= box.Y || visTop >= box.Y+box.H {
			s.p.SetBox(Rect{}) // off-screen
			continue
		}
		if visTop < box.Y {
			visTop = box.Y
		}
		if visBottom > box.Y+box.H {
			visBottom = box.Y + box.H
		}
		s.p.SetBox(Rect{X: box.X, Y: visTop, W: box.W, H: visBottom - visTop})
	}
}

func (g *Group) pageAffordance(m theme.Mode, pos, total int, up bool) string {
	glyph := "▼ more below"
	if up {
		glyph = "▲ more above"
	}
	txt := fmt.Sprintf("  %s · page %d/%d", glyph, pos, total)
	if cap, ok := m.Role("caption"); ok {
		return cap.Style().Render(txt)
	}
	return txt
}
