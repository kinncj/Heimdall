package pane

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/tui/theme"
)

// Group lays panes out in rows (one or more panes side by side per row) and adds
// Tab focus, a focus ring, two-level page+pane scroll, and pointer-targeted mouse.
//
// Focus is one-dimensional: Tab/Shift-Tab walk the panes in reading order
// (row-major), so a wide multi-column layout keeps a simple, predictable focus
// path. Layout is two-dimensional; focus is not.
//
// Two-level scroll: each pane scrolls its own overflow inside its box; the Group
// page-scrolls the whole stack when it is taller than the viewport. Both use the
// same keepVisible rule, so the focused pane's active row is never hidden.
type Group struct {
	rows       [][]*Pane
	focus      int // index into the row-major flattening
	pageScroll int
	box        Rect
	lastHeight int
}

// NewGroup builds a single-column group: one pane per row, stacked vertically.
func NewGroup(panes ...*Pane) *Group {
	rows := make([][]*Pane, len(panes))
	for i, p := range panes {
		rows[i] = []*Pane{p}
	}
	return NewGrid(rows)
}

// NewGrid builds a group from explicit rows; each row is one or more panes laid
// side by side. The first pane (row 0, col 0) starts focused.
func NewGrid(rows [][]*Pane) *Group {
	g := &Group{rows: rows}
	g.applyFocus()
	return g
}

// flat returns the panes in reading order (row-major).
func (g *Group) flat() []*Pane {
	var out []*Pane
	for _, row := range g.rows {
		out = append(out, row...)
	}
	return out
}

// Panes returns the panes in reading order.
func (g *Group) Panes() []*Pane { return g.flat() }

// Focus is the reading-order index of the focused pane.
func (g *Group) Focus() int { return g.focus }

// Focused returns the focused pane, or nil if the group is empty.
func (g *Group) Focused() *Pane {
	fl := g.flat()
	if len(fl) == 0 {
		return nil
	}
	return fl[g.focus]
}

// PageScroll is the current page (canvas) scroll offset in rows.
func (g *Group) PageScroll() int { return g.pageScroll }

func (g *Group) applyFocus() {
	i := 0
	for _, row := range g.rows {
		for _, p := range row {
			p.SetFocused(i == g.focus)
			i++
		}
	}
}

// FocusNext / FocusPrev cycle focus with wrap-around and page-scroll the canvas so
// the newly focused pane's active row comes into view.
func (g *Group) FocusNext() { g.setFocus(g.focus + 1) }
func (g *Group) FocusPrev() { g.setFocus(g.focus - 1) }

func (g *Group) setFocus(i int) {
	n := len(g.flat())
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
	handled := fp.Update(msg, g.innerHeight(height))
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
	fl := g.flat()
	inner := g.innerHeight(height)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		fl[idx].Wheel(-1, inner)
		if idx == g.focus {
			g.ensureFocusedVisible(height)
		}
		return true
	case tea.MouseButtonWheelDown:
		fl[idx].Wheel(1, inner)
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

// paneAt returns the reading-order index of the pane whose last-rendered box
// contains (x,y), or -1.
func (g *Group) paneAt(x, y int) int {
	for i, p := range g.flat() {
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
	flat  int // reading-order index
	x     int // column x offset within the group box
	w     int // column width
	top   int // canvas row of the row's top border
	inner int // content rows allocated to the pane's row
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

// layout places every pane on the canvas: rows stack vertically, panes within a
// row divide the width into equal columns. A row's height is the tallest pane in
// it, capped at one viewport so a giant pane scrolls internally.
func (g *Group) layout(width, height int) ([]slot, int) {
	maxInner := g.innerHeight(height)
	var slots []slot
	top, flat := 0, 0
	for _, row := range g.rows {
		ncols := len(row)
		if ncols == 0 {
			continue
		}
		colW := width / ncols
		if colW < 1 {
			colW = 1
		}
		rowInner := 1
		for _, p := range row {
			n := len(p.src.Rows())
			if n > maxInner {
				n = maxInner
			}
			if n > rowInner {
				rowInner = n
			}
		}
		for c, p := range row {
			x := c * colW
			w := colW
			if c == ncols-1 {
				w = width - x // last column takes the remainder
			}
			slots = append(slots, slot{p: p, flat: flat, x: x, w: w, top: top, inner: rowInner})
			flat++
		}
		top += rowInner + paneChrome
	}
	return slots, top
}

// ensureFocusedVisible page-scrolls the canvas so the focused pane's active row is
// on screen. Same keepVisible rule the panes use for their own cursors.
func (g *Group) ensureFocusedVisible(height int) {
	if height < 1 || len(g.flat()) == 0 {
		return
	}
	w := g.box.W
	if w < 1 {
		w = 1
	}
	slots, total := g.layout(w, height)
	s := slots[g.focus]
	within := s.p.ActiveRow() - s.p.Scroll()
	if within < 0 {
		within = 0
	}
	if within > s.inner-1 {
		within = s.inner - 1
	}
	canvasRow := s.top + 2 + within // +2: top border + title row
	g.pageScroll = clampOffset(keepVisible(g.pageScroll, canvasRow, height), total, height)
}

// --- rendering ---------------------------------------------------------------

// canvasSpan is one pane's column extent on a canvas line.
type canvasSpan struct{ flat, x, w int }

// canvasLine is one rendered row of the virtual canvas plus the panes it covers.
type canvasLine struct {
	line  string
	spans []canvasSpan
}

// View renders the rows into a height-row window at the given origin, page-scrolled
// so the focused pane's active row is visible, with ▲/▼ page affordances on the
// edges when the canvas overflows. box origin lets the mouse map screen coordinates
// back to panes.
func (g *Group) View(m theme.Mode, box Rect) string {
	g.box = box
	g.lastHeight = box.H
	if len(g.flat()) == 0 {
		return ""
	}
	slots, _ := g.layout(box.W, box.H)
	g.ensureFocusedVisible(box.H)

	// Render each row (its panes joined horizontally) into canvas lines, tagging
	// each line's column spans so we can set on-screen boxes after page-scrolling.
	byRow := map[int][]slot{}
	var order []int
	for _, s := range slots {
		if _, ok := byRow[s.top]; !ok {
			order = append(order, s.top)
		}
		byRow[s.top] = append(byRow[s.top], s)
	}
	var canvas []canvasLine
	for _, top := range order {
		rowSlots := byRow[top]
		cols := make([]string, len(rowSlots))
		spans := make([]canvasSpan, len(rowSlots))
		for i, s := range rowSlots {
			cols[i] = s.p.Frame(m, s.p.View(m, s.w-2, s.inner), s.w-2)
			spans[i] = canvasSpan{flat: s.flat, x: box.X + s.x, w: s.w}
		}
		block := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
		for _, ln := range strings.Split(block, "\n") {
			canvas = append(canvas, canvasLine{line: ln, spans: spans})
		}
	}

	realTotal := len(canvas)
	offset := clampOffset(g.pageScroll, realTotal, box.H)
	g.pageScroll = offset
	end := offset + box.H
	if end > realTotal {
		end = realTotal
	}

	g.setBoxes(canvas, offset, end, box)

	out := make([]string, 0, box.H)
	for _, cl := range canvas[offset:end] {
		out = append(out, cl.line)
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

// setBoxes assigns each pane its on-screen rectangle from the visible canvas rows,
// so the mouse can target it. A pane with no visible rows gets a zero box (no hits).
func (g *Group) setBoxes(canvas []canvasLine, offset, end int, box Rect) {
	type ext struct{ minY, maxY, x, w int }
	seen := map[int]*ext{}
	for j := offset; j < end; j++ {
		screenY := box.Y + (j - offset)
		for _, sp := range canvas[j].spans {
			e, ok := seen[sp.flat]
			if !ok {
				seen[sp.flat] = &ext{minY: screenY, maxY: screenY, x: sp.x, w: sp.w}
				continue
			}
			if screenY < e.minY {
				e.minY = screenY
			}
			if screenY > e.maxY {
				e.maxY = screenY
			}
		}
	}
	for i, p := range g.flat() {
		if e, ok := seen[i]; ok {
			p.SetBox(Rect{X: e.x, Y: e.minY, W: e.w, H: e.maxY - e.minY + 1})
		} else {
			p.SetBox(Rect{})
		}
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
