// Package pane is the one focus + scroll primitive for the Heimdall TUI
// (codename Himinbjörg). It replaces the three divergent scroll implementations
// that used to live in dashboard and topview.
//
// A Pane is one focusable, scrollable region. A Group (see group.go) stacks panes
// and adds Tab focus, a focus ring, two-level page+pane scroll, and mouse routing.
//
// Panes render only — they never classify or reshape metric meaning (ADR-0022).
// A pane's Source decides what its rows are and what filter/sort mean; the pane
// owns scrolling, selection, the "/" filter UX, the sort-key cycle, and the
// affordances. Capabilities are opt-in interfaces so a static panel gets neither
// filter nor sort.
package pane

import (
	"fmt"
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

// Filterable is a Source that supports the "/" full-text/name filter. What the
// query matches is the source's business (full-text for logs, name for commands).
type Filterable interface {
	SetFilter(query string)
}

// Sortable is a Source that supports an in-pane sort-key cycle. SortKeys lists the
// keys in cycle order; SetSort selects one.
type Sortable interface {
	SortKeys() []string
	SetSort(key string)
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

	filtering bool
	query     string
	sortIdx   int // index into SortKeys for Sortable sources

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

// Filtering reports whether the "/" filter input is open.
func (p *Pane) Filtering() bool { return p.filtering }

// CanFilter / CanSort report which capability affordances the footer should show.
func (p *Pane) CanFilter() bool { _, ok := p.src.(Filterable); return ok }
func (p *Pane) CanSort() bool   { _, ok := p.src.(Sortable); return ok }

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
// A pane in filter mode consumes typing; esc closes the filter. Otherwise the
// scroll/select/sort keys apply and everything else is left for the caller.
func (p *Pane) Update(msg tea.KeyMsg, height int) (handled bool) {
	if p.filtering {
		return p.updateFilter(msg)
	}
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
	case "/":
		if p.CanFilter() {
			p.filtering, p.query = true, ""
			return true
		}
		return false
	case "s":
		if p.CanSort() {
			p.cycleSort()
			return true
		}
		return false
	default:
		return false
	}
	return true
}

func (p *Pane) updateFilter(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "esc":
		p.filtering, p.query = false, ""
		if f, ok := p.src.(Filterable); ok {
			f.SetFilter("")
		}
		p.clampAfterChange(0)
	case "enter":
		p.filtering = false
	case "backspace":
		if r := []rune(p.query); len(r) > 0 {
			p.query = string(r[:len(r)-1])
			p.applyFilter()
		}
	default:
		if s := msg.String(); len(s) == 1 {
			p.query += s
			p.applyFilter()
		}
	}
	return true
}

func (p *Pane) applyFilter() {
	if f, ok := p.src.(Filterable); ok {
		f.SetFilter(p.query)
	}
	// A narrowed list resets the cursor to the top and re-clamps.
	if p.sel >= 0 {
		p.sel = 0
	}
	p.scroll = 0
}

func (p *Pane) cycleSort() {
	s, ok := p.src.(Sortable)
	if !ok {
		return
	}
	keys := s.SortKeys()
	if len(keys) == 0 {
		return
	}
	p.sortIdx = (p.sortIdx + 1) % len(keys)
	s.SetSort(keys[p.sortIdx])
}

// SortKey is the active sort key, or "" if the source is not Sortable.
func (p *Pane) SortKey() string {
	s, ok := p.src.(Sortable)
	if !ok {
		return ""
	}
	keys := s.SortKeys()
	if len(keys) == 0 {
		return ""
	}
	return keys[p.sortIdx%len(keys)]
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

func (p *Pane) clampAfterChange(height int) {
	if height < 1 {
		height = 1
	}
	p.scroll = clampOffset(p.scroll, p.rowCount(), height)
	if p.sel >= 0 {
		if p.sel > p.rowCount()-1 {
			p.sel = p.rowCount() - 1
		}
		if p.sel < 0 {
			p.sel = 0
		}
	}
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

	// Reserve the top row for the filter input when filtering.
	body := rows
	var head string
	if p.filtering {
		head = p.filterLine(m, width)
		height--
		if height < 1 {
			height = 1
		}
	}

	windowed := p.window(m, body, height)
	if head != "" {
		windowed = append([]string{head}, windowed...)
	}
	return strings.Join(windowed, "\n")
}

// window slices lines to a height-row window around p.scroll and (for a focused
// selectable pane) marks the cursor row, replacing the edge rows with themed
// "▲/▼ more · y/total" affordances so the height stays bounded.
func (p *Pane) window(m theme.Mode, lines []string, height int) []string {
	total := len(lines)
	offset := clampOffset(p.scroll, total, height)
	// Keep the cursor visible if this is a selectable pane.
	if p.sel >= 0 {
		offset = clampOffset(keepVisible(offset, p.sel, height), total, height)
	}
	p.scroll = offset

	if total <= height {
		return p.markCursor(m, append([]string(nil), lines...), offset)
	}

	out := append([]string(nil), lines[offset:offset+height]...)
	out = p.markCursor(m, out, offset)
	if offset > 0 {
		out[0] = p.affordance(m, offset, offset+1, total, true)
	}
	if offset+height < total {
		below := total - (offset + height)
		out[len(out)-1] = p.affordance(m, below, offset+height, total, false)
	}
	return out
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

func (p *Pane) affordance(m theme.Mode, n, pos, total int, up bool) string {
	glyph := "▼ more"
	if up {
		glyph = "▲ more"
	}
	txt := fmt.Sprintf("  %s · %d/%d", glyph, pos, total)
	if cap, ok := m.Role("caption"); ok {
		return cap.Style().Render(txt)
	}
	return txt
}

func (p *Pane) filterLine(m theme.Mode, width int) string {
	prompt := "/"
	if f, ok := m.Role("focus"); ok {
		prompt = f.Style().Render("/")
	}
	txt := prompt + p.query + "▏"
	if width > 0 {
		return lipgloss.NewStyle().MaxWidth(width).Render(txt)
	}
	return txt
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

// TransferStateFrom copies focus/scroll/selection/filter/sort from a same-shaped
// old group, so rebuilding on a live tick does not jump the view.
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
		np[i].filtering = op[i].filtering
		np[i].query = op[i].query
		np[i].sortIdx = op[i].sortIdx
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
