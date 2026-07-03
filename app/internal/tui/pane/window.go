package pane

import (
	"fmt"

	"heimdall/app/internal/tui/theme"
)

// Window renders lines into a height-row viewport around offset and returns the
// windowed lines plus the clamped offset. When sel >= 0 the window is nudged so
// that row stays visible, one row clear of the affordance edges. The scrolled
// edges are replaced with themed "▲/▼ more · pos/total" markers.
//
// This is the one windowing routine for the whole TUI (Himinbjörg): pane widgets
// use it via the Pane/Group, and simpler line-based views (the dashboard modals and
// detail body) call it directly instead of carrying their own copy.
func Window(m theme.Mode, lines []string, offset, sel, height int) ([]string, int) {
	total := len(lines)
	if height < 1 {
		height = 1
	}
	if total <= height {
		return append([]string(nil), lines...), 0
	}
	// Keep the selection clear of the row an affordance will overwrite, so the
	// cursor never hides under a "▲/▼ more" marker.
	if sel >= 0 {
		if sel <= offset {
			offset = sel - 1
		} else if sel >= offset+height-1 {
			offset = sel - height + 2
		}
	}
	offset = clampOffset(offset, total, height)
	out := append([]string(nil), lines[offset:offset+height]...)
	if offset > 0 {
		out[0] = affordanceLine(m, offset+1, total, true)
	}
	if offset+height < total {
		out[len(out)-1] = affordanceLine(m, offset+height, total, false)
	}
	return out, offset
}

// affordanceLine renders the "▲/▼ more · pos/total" edge marker in the caption role.
func affordanceLine(m theme.Mode, pos, total int, up bool) string {
	glyph := "▼ more"
	if up {
		glyph = "▲ more"
	}
	txt := fmt.Sprintf("  %s · %d/%d", glyph, pos, total)
	if c, ok := m.Role("caption"); ok {
		return c.Style().Render(txt)
	}
	return txt
}
