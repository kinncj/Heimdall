// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package brand renders Heimdall's TUI chrome — the header, status bar and
// splash — from the theme, matching assets/TUI_HEADER_* and TUI_STATUS_BAR.
// Images (Kitty/iTerm/Sixel) are a future enhancement; this renders the
// portable ASCII/box-drawing form that works on every terminal.
package brand

import (
	_ "embed"
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/tui/theme"
)

//go:embed assets/ASCII_ART.txt
var SplashArt string

//go:embed assets/ICON_ASCII_ART.txt
var IconArt string

// dashed is the header/status chrome border (matches the assets' dashed frame).
var dashed = lipgloss.Border{
	Top: "┄", Bottom: "┄", Left: "┊", Right: "┊",
	TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
}

func role(m theme.Mode, name string) lipgloss.Style {
	r, _ := m.Role(name)
	return r.Style()
}

func sep(m theme.Mode) string { return role(m, "border").Render(" │ ") }

func wordmark(m theme.Mode) string {
	t, _ := m.Role("title")
	return t.Style().Render(t.Glyph + " HEIMDALL")
}

func onlineBadge(m theme.Mode, online, total int) string {
	s, _ := m.State("online")
	return s.Style().Render(fmt.Sprintf("%s %d/%d ONLINE", s.Glyph, online, total))
}

func clockSeg(m theme.Mode, clock string) string {
	return role(m, "text_muted").Render("🕐 " + clock)
}

func chrome(m theme.Mode, width int, content string) string {
	bf, _ := m.Role("border")
	box := lipgloss.NewStyle().
		Border(dashed).
		BorderForeground(lipgloss.Color(bf.FG)).
		Padding(0, 1)
	if width > 4 {
		box = box.Width(width - 2)
	}
	return box.Render(content)
}

// SkinnyHeader is the default dashboard header: wordmark · online count · clock.
func SkinnyHeader(m theme.Mode, width, online, total int, clock string) string {
	content := wordmark(m) + sep(m) + onlineBadge(m, online, total) + sep(m) + clockSeg(m, clock)
	return chrome(m, width, content)
}

// FatHeader adds the tagline rule under the wordmark (splash / wide terminals).
func FatHeader(m theme.Mode, width, online, total int, clock string) string {
	a, _ := m.Role("accent")
	line1 := wordmark(m) + "    " + onlineBadge(m, online, total)
	line2 := a.Style().Render("-- watch over all realms --") + "    " + clockSeg(m, clock)
	return chrome(m, width, line1+"\n"+line2)
}

// StatusBar renders the footer chrome: brand · connection · source · transport ·
// clock. It reflects real state — whether the dashboard is currently receiving
// from its hub, and which hub (or "demo") it is bound to.
func StatusBar(m theme.Mode, width int, connected bool, source, clock string) string {
	conn, connRole := "○ connecting", "stale"
	if connected {
		conn, connRole = "● live", "online"
	}
	cs, _ := m.State(connRole)
	segs := []string{wordmark(m), cs.Style().Render(conn)}
	if source != "" {
		segs = append(segs, role(m, "text_muted").Render(source))
	}
	segs = append(segs, role(m, "text_muted").Render("gRPC"), clockSeg(m, clock))

	content := segs[0]
	for _, s := range segs[1:] {
		content += sep(m) + s
	}
	return chrome(m, width, content)
}
