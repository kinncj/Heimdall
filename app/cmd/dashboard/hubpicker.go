// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/discovery"
	"heimdall/app/internal/tui/theme"
)

// pickHub presents the discovered hubs and returns the address the operator
// chooses. Used when zeroconf finds more than one hub on the LAN (v2).
func pickHub(hubs []discovery.Found, md theme.Mode) (string, error) {
	res, err := tea.NewProgram(hubPicker{hubs: hubs, mode: md}, tea.WithAltScreen()).Run()
	if err != nil {
		return "", err
	}
	hp, _ := res.(hubPicker)
	if hp.chosen == "" {
		return "", fmt.Errorf("no hub selected")
	}
	return hp.chosen, nil
}

type hubPicker struct {
	hubs   []discovery.Found
	sel    int
	chosen string
	mode   theme.Mode
}

func (m hubPicker) Init() tea.Cmd { return nil }

func (m hubPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if m.sel < len(m.hubs)-1 {
				m.sel++
			}
		case "enter":
			m.chosen = m.hubs[m.sel].Addr
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m hubPicker) View() string {
	heading, _ := m.mode.Role("heading")
	val, _ := m.mode.Role("value")
	focus, _ := m.mode.Role("focus")
	muted, _ := m.mode.Role("text_muted")

	var b strings.Builder
	b.WriteString("\n  " + heading.Style().Render("Multiple Heimdall hubs found — pick one") + "\n\n")
	for i, h := range m.hubs {
		name := h.Name
		if name == "" {
			name = h.Addr
		}
		line := fmt.Sprintf("%-24s %s", name, h.Addr)
		if i == m.sel {
			b.WriteString("  " + focus.Style().Render("▸ "+line) + "\n")
		} else {
			b.WriteString("    " + val.Style().Render(line) + "\n")
		}
	}
	b.WriteString("\n  " + muted.Style().Render("↑/↓ pick   ⏎ connect   q quit") + "\n")
	return b.String()
}
