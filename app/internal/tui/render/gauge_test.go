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
	got := strip(Sparkline(m, []float64{0, 25, 50, 75, 100}, 0))
	if len([]rune(got)) != 5 {
		t.Errorf("sparkline len = %d, want 5 (%q)", len([]rune(got)), got)
	}
}

func TestSparklineClampsToWidth(t *testing.T) {
	m := darkMode(t)
	history := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	got := strip(Sparkline(m, history, 4))
	if n := len([]rune(got)); n != 4 {
		t.Fatalf("clamped sparkline len = %d, want 4 (%q)", n, got)
	}
	// Must keep the most recent samples (the tail), not the head.
	if want := strip(Sparkline(m, history[len(history)-4:], 0)); got != want {
		t.Errorf("clamped sparkline = %q, want tail %q", got, want)
	}
}

func TestBrailleSparklineRampAndBounds(t *testing.T) {
	m := darkMode(t)
	got := []rune(strip(BrailleSparkline(m, []float64{0, 50, 100}, 0)))
	if len(got) != 3 {
		t.Fatalf("braille len = %d, want 3 (%q)", len(got), string(got))
	}
	lo, hi := got[0], got[2]
	// Lowest sample uses the lowest ramp glyph, highest uses the fullest.
	if i := runeIndex(brailleRunes, lo); i != 0 {
		t.Errorf("min sample glyph %q at ramp index %d, want 0", lo, i)
	}
	if i := runeIndex(brailleRunes, hi); i != len(brailleRunes)-1 {
		t.Errorf("max sample glyph %q at ramp index %d, want top", hi, i)
	}
	// All glyphs must come from the braille ramp (display-width 1, NO_COLOR safe).
	for _, r := range got {
		if runeIndex(brailleRunes, r) < 0 {
			t.Errorf("glyph %q is not in the braille ramp", r)
		}
	}
}

func TestBrailleSparklineEmptyAndClamp(t *testing.T) {
	m := darkMode(t)
	if got := BrailleSparkline(m, nil, 8); got != "" {
		t.Errorf("empty history should render empty, got %q", got)
	}
	history := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	if n := len([]rune(strip(BrailleSparkline(m, history, 4)))); n != 4 {
		t.Errorf("clamped braille len = %d, want 4", n)
	}
}

func runeIndex(rs []rune, r rune) int {
	for i, x := range rs {
		if x == r {
			return i
		}
	}
	return -1
}
