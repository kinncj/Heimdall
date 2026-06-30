// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// refresh re-collects metrics immediately — the same work a tick does — so the
// operator can force an update with 'r' without waiting for the next tick.
func (m Model) refresh() {
	if m.onTick != nil {
		m.onTick(m.now)
	}
	m.reg.Evaluate(m.now)
	m.recordHistory()
}

// keyBindings is the help overlay's content. An empty key marks a section header.
var keyBindings = []struct{ key, desc string }{
	{"", "Fleet"},
	{"↑/↓ · j/k", "move selection"},
	{"⏎", "open host detail"},
	{"g", "cycle grouping (hub / os / tag)"},
	{"/", "filter (host, tag, hub, os, state)"},
	{"r", "refresh now"},
	{"", "Host detail"},
	{"↑/↓", "previous / next host"},
	{"⇧↑/↓ · wheel", "scroll the detail"},
	{"l / p / c", "logs / processes / command (when offered)"},
	{"t", "full-screen top view"},
	{"", "Modals"},
	{"↑/↓ · wheel", "pick / scroll"},
	{"⏎", "open / run"},
	{"/", "search logs"},
	{"s", "change process sort"},
	{"esc", "back — one level at a time"},
	{"", "General"},
	{"?", "toggle this help"},
	{"q · ctrl+c", "quit"},
}

// HelpView renders the key-binding overlay, centered on the frame.
func (m Model) HelpView() string {
	heading, _ := m.mode.Role("heading")
	keys, _ := m.mode.Role("keybinding")
	label, _ := m.mode.Role("label")
	muted, _ := m.mode.Role("text_muted")

	var b strings.Builder
	b.WriteString(heading.Style().Render("⬢ HEIMDALL — Key Bindings") + "\n")
	for _, kb := range keyBindings {
		if kb.key == "" {
			b.WriteString("\n" + label.Style().Render(kb.desc) + "\n")
			continue
		}
		b.WriteString("  " + keys.Style().Render(fmt.Sprintf("%-13s", kb.key)) +
			muted.Style().Render(kb.desc) + "\n")
	}
	b.WriteString("\n" + muted.Style().Render("Watch Over All Realms"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 3)
	if fg := heading.Style().GetForeground(); fg != nil {
		box = box.BorderForeground(fg)
	}
	panel := box.Render(b.String())

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}
	return panel
}
