// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package theme

import "testing"

func hasAttr(a []string, x string) bool {
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}

func TestLoadDefaultDark(t *testing.T) {
	th, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if th.DefaultMode != "dark" {
		t.Fatalf("default mode %q, want dark", th.DefaultMode)
	}
	if _, ok := th.Mode(""); !ok {
		t.Fatal("default mode not resolvable")
	}
}

func TestTitleIsSteelBold(t *testing.T) {
	th, _ := Load()
	m, _ := th.Mode("dark")
	ti, ok := m.Role("title")
	if !ok {
		t.Fatal("no title role")
	}
	if ti.FG != "#d0d0d0" {
		t.Errorf("title fg %s, want steel #d0d0d0", ti.FG)
	}
	if !hasAttr(ti.Attrs, "bold") {
		t.Errorf("title should be bold, attrs=%v", ti.Attrs)
	}
	if ti.Glyph != "⬢" {
		t.Errorf("title glyph %q, want ⬢", ti.Glyph)
	}
}

func TestAccentIsElectricBlue(t *testing.T) {
	th, _ := Load()
	m, _ := th.Mode("dark")
	a, _ := m.Role("accent")
	if a.FG != "#00d7ff" {
		t.Errorf("accent fg %s, want electric-blue #00d7ff", a.FG)
	}
}

func TestStateOnline(t *testing.T) {
	th, _ := Load()
	m, _ := th.Mode("dark")
	s, ok := m.State("online")
	if !ok {
		t.Fatal("no online state")
	}
	if s.FG != "#5fd75f" || s.Glyph != "●" || s.Label != "ONLINE" {
		t.Errorf("online = %+v, want #5fd75f ● ONLINE", s)
	}
}

func TestNeedsHelperIsInfoCyanNeverRed(t *testing.T) {
	th, _ := Load()
	m, _ := th.Mode("dark")
	s, _ := m.State("needs_helper")
	if s.FG != "#5fd7ff" {
		t.Errorf("needs_helper fg %s, want info-cyan #5fd7ff (never red)", s.FG)
	}
}

func TestSeverityForPercent(t *testing.T) {
	th, _ := Load()
	m, _ := th.Mode("dark")
	for _, c := range []struct {
		pct  float64
		name string
		fg   string
	}{
		{10, "nominal", "#5fd7af"},
		{50, "moderate", "#87d75f"},
		{65, "elevated", "#ffd75f"},
		{80, "high", "#ff875f"},
		{95, "critical", "#ff5f5f"},
	} {
		tier, name := m.SeverityFor(c.pct)
		if name != c.name || tier.FG != c.fg {
			t.Errorf("SeverityFor(%.0f) = %s/%s, want %s/%s", c.pct, name, tier.FG, c.name, c.fg)
		}
	}
}
