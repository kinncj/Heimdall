// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package render

import (
	"regexp"
	"strings"
	"testing"

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
	m, _ := th.Mode("dark")
	return m
}

func TestFilledCells(t *testing.T) {
	cases := []struct {
		pct   float64
		cells int
		want  int
	}{{0, 10, 0}, {50, 10, 5}, {100, 10, 10}, {95, 20, 19}, {-5, 10, 0}, {150, 10, 10}}
	for _, c := range cases {
		if got := FilledCells(c.pct, c.cells); got != c.want {
			t.Errorf("FilledCells(%.0f,%d)=%d want %d", c.pct, c.cells, got, c.want)
		}
	}
}

func TestGaugeRendersBlocks(t *testing.T) {
	m := darkMode(t)
	got := strip(Gauge(m, 50, 10))
	if fill := strings.Count(got, "█"); fill != 5 {
		t.Errorf("gauge 50%% -> %d filled blocks, want 5 (%q)", fill, got)
	}
	if track := strings.Count(got, "░"); track != 5 {
		t.Errorf("gauge 50%% -> %d track blocks, want 5 (%q)", track, got)
	}
}

func TestGaugeExtremes(t *testing.T) {
	m := darkMode(t)
	if got := strip(Gauge(m, 0, 8)); got != strings.Repeat("░", 8) {
		t.Errorf("0%% gauge = %q", got)
	}
	if got := strip(Gauge(m, 100, 8)); got != strings.Repeat("█", 8) {
		t.Errorf("100%% gauge = %q", got)
	}
}

func TestSparklineLength(t *testing.T) {
	m := darkMode(t)
	got := strip(Sparkline(m, []float64{0, 25, 50, 75, 100}))
	if len([]rune(got)) != 5 {
		t.Errorf("sparkline len = %d, want 5 (%q)", len([]rune(got)), got)
	}
}
