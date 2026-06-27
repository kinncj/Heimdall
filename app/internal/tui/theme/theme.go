// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package theme loads the Heimdall terminal theme (a copy of the approved
// docs/design/identity/terminal-theme.json, embedded so runtime never imports
// from docs/) and exposes it as lipgloss styles.
package theme

import (
	_ "embed"
	"encoding/json"

	"github.com/charmbracelet/lipgloss"
)

//go:embed heimdall-theme.json
var themeJSON []byte

// RoleStyle is a structural role (fg and/or bg + attributes + optional glyph).
type RoleStyle struct {
	FG    string   `json:"fg"`
	BG    string   `json:"bg"`
	Ansi  int      `json:"ansi256"`
	Attrs []string `json:"attrs"`
	Glyph string   `json:"glyph"`
}

// StateStyle is a status role: colour + glyph + word + emphasis.
type StateStyle struct {
	FG    string   `json:"fg"`
	Ansi  int      `json:"ansi256"`
	Glyph string   `json:"glyph"`
	Label string   `json:"label"`
	Attrs []string `json:"attrs"`
}

// SeverityTier is one stop of the gauge-fill ramp.
type SeverityTier struct {
	FG   string `json:"fg"`
	Ansi int    `json:"ansi256"`
	Band string `json:"band"`
}

// Mode is one theme variant (dark or light).
type Mode struct {
	Structure map[string]RoleStyle    `json:"structure"`
	States    map[string]StateStyle   `json:"states"`
	Severity  map[string]SeverityTier `json:"severity"`
}

// Theme is the parsed terminal theme.
type Theme struct {
	Name        string          `json:"name"`
	DefaultMode string          `json:"default_mode"`
	Modes       map[string]Mode `json:"modes"`
}

// Load parses the embedded Heimdall theme.
func Load() (*Theme, error) {
	var t Theme
	if err := json.Unmarshal(themeJSON, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Mode returns a mode by name; "" resolves to the default.
func (t *Theme) Mode(name string) (Mode, bool) {
	if name == "" {
		name = t.DefaultMode
	}
	m, ok := t.Modes[name]
	return m, ok
}

// Role returns a structural role by name.
func (m Mode) Role(name string) (RoleStyle, bool) {
	r, ok := m.Structure[name]
	return r, ok
}

// State returns a status role by name.
func (m Mode) State(name string) (StateStyle, bool) {
	s, ok := m.States[name]
	return s, ok
}

// SeverityFor maps a 0–100 percentage to its gauge-fill tier and name.
func (m Mode) SeverityFor(pct float64) (SeverityTier, string) {
	var name string
	switch {
	case pct < 40:
		name = "nominal"
	case pct < 60:
		name = "moderate"
	case pct < 75:
		name = "elevated"
	case pct < 90:
		name = "high"
	default:
		name = "critical"
	}
	return m.Severity[name], name
}

func applyAttrs(s lipgloss.Style, attrs []string) lipgloss.Style {
	for _, a := range attrs {
		switch a {
		case "bold":
			s = s.Bold(true)
		case "faint":
			s = s.Faint(true)
		case "italic":
			s = s.Italic(true)
		case "underline":
			s = s.Underline(true)
		case "reverse":
			s = s.Reverse(true)
		}
	}
	return s
}

// Style builds a lipgloss style for a structural role. lipgloss renders
// truecolor and auto-degrades to ANSI-256/16 / NO_COLOR per the terminal.
func (r RoleStyle) Style() lipgloss.Style {
	s := lipgloss.NewStyle()
	if r.FG != "" {
		s = s.Foreground(lipgloss.Color(r.FG))
	}
	if r.BG != "" {
		s = s.Background(lipgloss.Color(r.BG))
	}
	return applyAttrs(s, r.Attrs)
}

// Style builds a lipgloss style for a status role.
func (s StateStyle) Style() lipgloss.Style {
	st := lipgloss.NewStyle()
	if s.FG != "" {
		st = st.Foreground(lipgloss.Color(s.FG))
	}
	return applyAttrs(st, s.Attrs)
}

// Badge renders "<glyph> <LABEL>" in the state's style.
func (s StateStyle) Badge() string {
	return s.Style().Render(s.Glyph + " " + s.Label)
}
