// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/theme"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func darkMode(t *testing.T) theme.Mode {
	t.Helper()
	th, err := theme.Load()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := th.Mode("dark")
	if !ok {
		t.Fatal("no dark mode")
	}
	return m
}

func key(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func TestHelpToggleAndView(t *testing.T) {
	reg := domain.NewHostRegistry(2*time.Second, 5*time.Second)
	m := New(darkMode(t), reg, time.Now())

	if m.help {
		t.Fatal("help should start closed")
	}
	m2, _ := m.Update(key('?'))
	m = m2.(Model)
	if !m.help {
		t.Fatal("'?' should open help")
	}

	frame := ansiRe.ReplaceAllString(m.View(), "")
	// The help must document the current bindings, including the v2 ones, grouped
	// into sections — not just the basic grid keys (regression: help was stale).
	for _, want := range []string{
		"Key Bindings", "refresh now", "quit",
		"Fleet", "Host detail", "Modals",
		"grouping", "filter", "scroll", "logs", "top", "command", "sort",
	} {
		if !strings.Contains(frame, want) {
			t.Errorf("help frame missing %q", want)
		}
	}

	m2, _ = m.Update(key('?'))
	if m2.(Model).help {
		t.Fatal("'?' should close help when open")
	}
}

func TestRefreshKeyInvokesTick(t *testing.T) {
	reg := domain.NewHostRegistry(2*time.Second, 5*time.Second)
	ticks := 0
	m := New(darkMode(t), reg, time.Now()).WithTick(func(time.Time) { ticks++ })

	m2, _ := m.Update(key('r'))
	if m2.(Model).help {
		t.Fatal("'r' must not open help")
	}
	if ticks != 1 {
		t.Fatalf("refresh tick count = %d, want 1", ticks)
	}
}
