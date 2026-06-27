// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package brand

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/tui/theme"
)

var ansi = regexp.MustCompile("\x1b\\[[0-9;]*m")

func strip(s string) string { return ansi.ReplaceAllString(s, "") }

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

func TestSkinnyHeaderHasLiveElements(t *testing.T) {
	m := darkMode(t)
	out := strip(SkinnyHeader(m, 90, 6, 7, "14:03:12"))
	for _, want := range []string{"HEIMDALL", "6/7 ONLINE", "14:03:12"} {
		if !strings.Contains(out, want) {
			t.Errorf("skinny header missing %q\n%s", want, out)
		}
	}
}

func TestHeaderFitsWidth(t *testing.T) {
	m := darkMode(t)
	const w = 90
	for _, line := range strings.Split(SkinnyHeader(m, w, 6, 7, "14:03:12"), "\n") {
		if lipgloss.Width(line) > w {
			t.Errorf("header line exceeds width %d: got %d\n%q", w, lipgloss.Width(line), strip(line))
		}
	}
}

func TestStatusBarHasFields(t *testing.T) {
	m := darkMode(t)
	out := strip(StatusBar(m, 120, true, "localhost:9090", "14:03:12"))
	for _, want := range []string{"HEIMDALL", "live", "localhost:9090", "gRPC", "14:03:12"} {
		if !strings.Contains(out, want) {
			t.Errorf("status bar missing %q\n%s", want, out)
		}
	}
}

func TestFatHeaderHasTagline(t *testing.T) {
	m := darkMode(t)
	out := strip(FatHeader(m, 90, 6, 7, "14:03:12"))
	if !strings.Contains(out, "watch over all realms") {
		t.Errorf("fat header missing tagline\n%s", out)
	}
}
