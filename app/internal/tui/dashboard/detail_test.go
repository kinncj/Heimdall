// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/tui/theme"
)

func themed(t *testing.T) theme.Mode {
	t.Helper()
	th, err := theme.Load()
	if err != nil {
		t.Fatal(err)
	}
	mode, _ := th.Mode("dark")
	return mode
}

// The detail view must reflow to the terminal width like the top view, not
// overflow on a narrow terminal (e.g. SSH from a phone in Termius).
func TestDetailViewFitsEveryWidth(t *testing.T) {
	m := Model{reg: obsReg(t), mode: themed(t), detail: true, height: 40}
	for _, w := range []int{120, 80, 50, 36} {
		m.width = w
		for _, ln := range strings.Split(m.DetailView(), "\n") {
			if got := lipgloss.Width(ln); got > w {
				t.Errorf("width %d: line of %d cols exceeds it: %q", w, got, ln)
			}
		}
	}
}
