// Package pane is the one focus + scroll primitive for the Heimdall TUI
// (codename Himinbjörg). It replaces the three divergent scroll implementations
// that used to live in dashboard and topview.
//
// A Pane is one focusable, scrollable region. A Group (see group.go) stacks panes
// and adds Tab focus, a focus ring, two-level page+pane scroll, and mouse routing.
//
// Panes render only — they never classify or reshape metric meaning (ADR-0022).
// A pane's Source decides what its rows are; the pane owns scrolling, selection,
// the focus ring, and the "▲/▼ more" affordances. Filtering and sorting are the
// consumer's concern (e.g. the dashboard), not the pane's.
package pane

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/tui/theme"
)

// Source supplies the rows a pane displays. Rows returns already-styled lines in
// the current filter/sort order.
type Source interface {
	Rows() []string
}

// Selectable is a Source whose rows carry a moving cursor. RowCount is the number
// of selectable rows (== len(Rows)); a pane uses it to keep the cursor in view.
type Selectable interface {
	Source
	RowCount() int
}

// Rect is a pane's on-screen box, captured at render so the mouse can target it.
type Rect struct{ X, Y, W, H int }

// Contains reports whether the point (x,y) falls inside the box.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Pane is one focusable, scrollable region.
type Pane struct {
	title string
	src   Source

	scroll int // first visible row of the source
	sel    int // cursor row for Selectable sources; -1 otherwise

	box     Rect // set at View time; read for mouse hit-testing
	focused bool
}

// New builds a pane over src. If src is Selectable the pane starts with the cursor
// on the first row; otherwise it is a scroll-only pane.
func New(title string, src Source) *Pane {
	p := &Pane{title: title, src: src, sel: -1}
	if _, ok := src.(Selectable); ok {
		p.sel = 0
	}
	return p
}

// Title is the pane's heading.
func (p *Pane) Title() string { return p.title }

// Box returns the last-rendered on-screen rectangle (for mouse hit-testing).
func (p *Pane) Box() Rect { return p.box }

// SetBox records the pane's on-screen rectangle (called by Group at layout time).
func (p *Pane) SetBox(r Rect) { p.box = r }

// SetFocused toggles the focus ring / active-row caret / capability footer.
func (p *Pane) SetFocused(f bool) { p.focused = f }

// Focused reports whether the pane currently draws the focus ring.
func (p *Pane) Focused() bool { return p.focused }

// selectable returns the source as Selectable, or nil.
func (p *Pane) selectable() Selectable {
	s, _ := p.src.(Selectable)
	return s
}

// rowCount is the number of source rows (selectable count when Selectable, else
// the rendered line count).
func (p *Pane) rowCount() int {
	if s := p.selectable(); s != nil {
		return s.RowCount()
	}
	return len(p.src.Rows())
}

// --- pure scroll/selection math (no theme, unit-tested directly) -------------

// clampOffset clamps a scroll offset to [0, max(0, total-height)].
func clampOffset(offset, total, height int) int {
	if height < 1 {
		height = 1
	}
	max := total - height
	if max < 0 {
		max = 0
	}
	if offset > max {
		offset = max
	}
	if offset < 0 {
		offset = 0
	}
	return offset
}

// keepVisible nudges offset so that row `target` sits inside the height-row window
// starting at offset. It moves the window the minimum distance needed — the
// selection-follows-cursor and page-auto-scroll rule, in one place.
func keepVisible(offset, target, height int) int {
	if height < 1 {
		height = 1
	}
	if target < offset {
		return target
	}
	if target >= offset+height {
		return target - height + 1
	}
	return offset
}

// --- input -------------------------------------------------------------------

// Update handles a key for a focused pane and reports whether it consumed the key.
// Scroll/select keys apply; everything else is left for the caller.
func (p *Pane) Update(msg tea.KeyMsg, height int) (handled bool) {
	switch msg.String() {
	case "up", "k":
		p.move(-1, height)
	case "down", "j":
		p.move(1, height)
	case "pgup":
		p.move(-p.page(height), height)
	case "pgdown", "pgdn":
		p.move(p.page(height), height)
	case "home":
		p.moveTo(0, height)
	case "end":
		p.moveTo(p.rowCount()-1, height)
	default:
		return false
	}
	return true
}

// move shifts the cursor (selectable) or the scroll offset (scroll-only) by delta
// and keeps the result visible in a height-row window.
func (p *Pane) move(delta, height int) {
	if p.sel >= 0 {
		p.moveTo(p.sel+delta, height)
		return
	}
	p.scroll = clampOffset(p.scroll+delta, p.rowCount(), height)
}

func (p *Pane) moveTo(target, height int) {
	n := p.rowCount()
	if p.sel >= 0 {
		if target < 0 {
			target = 0
		}
		if target > n-1 {
			target = n - 1
		}
		if target < 0 {
			target = 0
		}
		p.sel = target
		p.scroll = clampOffset(keepVisible(p.scroll, p.sel, height), n, height)
		return
	}
	p.scroll = clampOffset(target, n, height)
}

// Wheel scrolls the pane by dir (-1 up, +1 down) worth of a few rows, as the mouse
// wheel does. It moves the cursor for selectable panes so the highlight follows.
func (p *Pane) Wheel(dir, height int) {
	p.move(dir*wheelStep, height)
}

const wheelStep = 3

// Selection is the current cursor row (or -1 for a scroll-only pane).
func (p *Pane) Selection() int { return p.sel }

// Scroll is the current top visible row.
func (p *Pane) Scroll() int { return p.scroll }

// ActiveRow is the row the pane most wants kept on screen — the cursor for a
// selectable pane, else the top visible row. Group uses it for page-auto-scroll.
func (p *Pane) ActiveRow() int {
	if p.sel >= 0 {
		return p.sel
	}
	return p.scroll
}

// --- rendering ---------------------------------------------------------------

// View renders the pane into height rows (its inner content area, before the
// border) at the given width, windowed around the scroll offset with the standard
// affordances. The focus ring (border weight) is applied by Group via Frame.
func (p *Pane) View(m theme.Mode, width, height int) string {
	rows := p.src.Rows()
	if len(rows) == 0 {
		muted, _ := m.Role("text_muted")
		rows = []string{muted.Style().Render("  (empty)")}
	}
	if height < 1 {
		height = 1
	}
	return strings.Join(p.window(m, rows, height), "\n")
}

// window renders lines through the shared Window (so scrolling, the selection
// margin, and the "▲/▼ more · y/total" affordances behave identically to the
// dashboard modals) and then marks the focused pane's cursor row. Window keeps the
// selection clear of the affordance edges, so the cursor is never overwritten.
func (p *Pane) window(m theme.Mode, lines []string, height int) []string {
	out, offset := Window(m, lines, p.scroll, p.sel, height)
	p.scroll = offset
	return p.markCursor(m, out, offset)
}

// markCursor prefixes the focused pane's cursor row with the ▸ caret (in the
// selection role) and other rows with a space, so focus survives NO_COLOR.
func (p *Pane) markCursor(m theme.Mode, win []string, offset int) []string {
	if p.sel < 0 || !p.focused {
		return win
	}
	sel, ok := m.Role("selection")
	for i := range win {
		row := offset + i
		if row == p.sel {
			caret := "▸ "
			if ok {
				caret = sel.Style().Render("▸") + " "
			}
			win[i] = caret + strings.TrimPrefix(win[i], "  ")
		}
	}
	return win
}

// Frame wraps content in a border whose weight signals focus: a heavy border when
// focused (the focus ring), a normal border otherwise. Colour comes from the focus
// / border role; the weight is the non-colour signal, so focus survives NO_COLOR.
func (p *Pane) Frame(m theme.Mode, content string, width int) string {
	border := lipgloss.NormalBorder()
	roleName := "border"
	if p.focused {
		border = lipgloss.ThickBorder()
		roleName = "focus"
	}
	st := lipgloss.NewStyle().Border(border).Padding(0, 1)
	if role, ok := m.Role(roleName); ok && role.FG != "" {
		st = st.BorderForeground(lipgloss.Color(role.FG))
	}
	if width > 0 {
		st = st.Width(width - 2) // Padding(0,1) adds two columns
	}
	title := p.title
	if h, ok := m.Role("heading"); ok {
		title = h.Style().Render(p.title)
	}
	return st.Render(title + "\n" + content)
}

// SetScroll / SetSelection restore view state (used when a live refresh rebuilds
// panes but must keep the user where they were).
func (p *Pane) SetScroll(v int) { p.scroll = v }
func (p *Pane) SetSelection(v int) {
	if p.sel >= 0 {
		p.sel = v
	}
}

// TransferStateFrom copies focus/scroll/selection from a same-shaped old group, so
// rebuilding on a live tick does not jump the view.
func (g *Group) TransferStateFrom(old *Group) {
	if old == nil {
		return
	}
	np, op := g.flat(), old.flat()
	if len(np) != len(op) {
		return
	}
	g.focus = old.focus
	g.pageScroll = old.pageScroll
	for i := range np {
		np[i].scroll = op[i].scroll
		np[i].sel = op[i].sel
	}
	g.applyFocus()
}

// page is one page of scrolling (a near-full window, keeping one row of overlap).
func (p *Pane) page(height int) int {
	if height > 1 {
		return height - 1
	}
	return 1
}
