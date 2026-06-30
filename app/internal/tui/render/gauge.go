// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package render holds pure TUI render helpers (gauges, sparklines) built on the
// theme. They take values + width and return styled strings — no Bubble Tea.
package render

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/tui/theme"
)

// FilledCells returns how many of cells are filled for a 0–100 percentage.
func FilledCells(pct float64, cells int) int {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	f := int(math.Round(pct / 100 * float64(cells)))
	if f > cells {
		f = cells
	}
	return f
}

// Gauge renders a btop-style fill bar: filled blocks in the severity-ramp colour
// for the value, the remainder as a muted track. Colour reinforces the value,
// which is also shown numerically by the caller, so it survives NO_COLOR.
func Gauge(m theme.Mode, pct float64, cells int) string {
	filled := FilledCells(pct, cells)
	tier, _ := m.SeverityFor(pct)
	fill := lipgloss.NewStyle().Foreground(lipgloss.Color(tier.FG)).Render(strings.Repeat("█", filled))
	tm, _ := m.Role("text_muted")
	track := tm.Style().Render(strings.Repeat("░", cells-filled))
	return fill + track
}

var sparkRunes = []rune("▁▂▃▄▅▆▇█")

// Sparkline renders the trailing window of a value history (each 0–100) as a
// braille-ish trend. width caps how many samples are drawn so the line never
// runs past its column; width <= 0 draws the whole history.
func Sparkline(m theme.Mode, history []float64, width int) string {
	if width > 0 && len(history) > width {
		history = history[len(history)-width:]
	}
	if len(history) == 0 {
		return ""
	}
	var b strings.Builder
	for _, v := range history {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		idx := int(math.Round(v / 100 * float64(len(sparkRunes)-1)))
		b.WriteRune(sparkRunes[idx])
	}
	s, _ := m.Role("text_secondary")
	return s.Style().Render(b.String())
}

// brailleRunes is a bottom-up braille fill ramp — the btop/mactop sparkline look.
// Each glyph is display-width 1.
var brailleRunes = []rune("⣀⣤⣶⣷⣿")

// BrailleSparkline renders the trailing window of a value history (each 0–100) as
// a braille trend line. The body is drawn in text_secondary; the most recent
// sample is tinted by its severity tier so the current level reads at a glance.
// The numeric value is shown beside it by the caller, so the trend survives
// NO_COLOR. width caps how many samples are drawn; width <= 0 draws all.
func BrailleSparkline(m theme.Mode, history []float64, width int) string {
	if width > 0 && len(history) > width {
		history = history[len(history)-width:]
	}
	if len(history) == 0 {
		return ""
	}
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 100 {
			return 100
		}
		return v
	}
	glyph := func(v float64) rune {
		idx := int(math.Round(clamp(v) / 100 * float64(len(brailleRunes)-1)))
		return brailleRunes[idx]
	}
	sec, _ := m.Role("text_secondary")
	if len(history) == 1 {
		tier, _ := m.SeverityFor(clamp(history[0]))
		return lipgloss.NewStyle().Foreground(lipgloss.Color(tier.FG)).Render(string(glyph(history[0])))
	}
	var b strings.Builder
	for _, v := range history[:len(history)-1] {
		b.WriteRune(glyph(v))
	}
	body := sec.Style().Render(b.String())
	last := history[len(history)-1]
	tier, _ := m.SeverityFor(clamp(last))
	tip := lipgloss.NewStyle().Foreground(lipgloss.Color(tier.FG)).Render(string(glyph(last)))
	return body + tip
}
