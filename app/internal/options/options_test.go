// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import (
	"flag"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sampleCatalog() Catalog {
	return NewCatalog(
		Define("hub").Default("localhost:9090").Help("hub").Ask("Hub?"),
		Define("tls").Of(KindToggle).Default("false").Help("tls"),
		Define("interval").Of(KindSpan).Default("2s").Help("interval"),
		Define("token").Of(KindSecret).Env("HEIMDALL_TEST_TOKEN").Help("token"),
	)
}

func TestPrecedenceFlagOverFileOverDefault(t *testing.T) {
	cat := sampleCatalog()

	// default only
	r := NewResolver(cat).With(Builtins(cat)).Resolve()
	if got := r.Address("hub").String(); got != "localhost:9090" {
		t.Fatalf("default hub = %q", got)
	}

	// file overrides default
	file := mapSource{values: map[Key]string{"hub": "file-host:1"}}
	r = NewResolver(cat).With(Builtins(cat)).With(file).Resolve()
	if got := r.Address("hub").String(); got != "file-host:1" {
		t.Fatalf("file hub = %q", got)
	}

	// an explicitly-set flag wins over the file
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	cat.Register(fs)
	_ = fs.Parse([]string{"--hub", "flag-host:2"})
	r = NewResolver(cat).With(Builtins(cat)).With(file).With(FromFlags(cat, fs)).Resolve()
	if got := r.Address("hub").String(); got != "flag-host:2" {
		t.Fatalf("flag hub should win, got %q", got)
	}
	if !AnyProvided(fs) {
		t.Fatal("AnyProvided should be true")
	}
}

func TestEnvSourceAndTypedAccessors(t *testing.T) {
	t.Setenv("HEIMDALL_TEST_TOKEN", "s3cret")
	cat := sampleCatalog()
	file := mapSource{values: map[Key]string{"tls": "true", "interval": "5s"}}
	r := NewResolver(cat).With(Builtins(cat)).With(file).With(FromEnvironment(cat)).Resolve()

	if !r.Toggle("tls") {
		t.Error("tls should be true from file")
	}
	if r.Span("interval", time.Second) != 5*time.Second {
		t.Errorf("interval = %s", r.Span("interval", time.Second))
	}
	sec := r.Secret("token")
	if sec.Reveal() != "s3cret" {
		t.Errorf("token reveal = %q", sec.Reveal())
	}
	if strings.Contains(sec.String(), "s3cret") {
		t.Errorf("secret String must redact, got %q", sec.String())
	}
}

func TestFileRoundTripTypesValues(t *testing.T) {
	cat := sampleCatalog()
	path := filepath.Join(t.TempDir(), "sub", "daemon.json")

	resolved := NewResolver(cat).With(Builtins(cat)).
		With(mapSource{values: map[Key]string{"hub": "h:9", "tls": "true", "token": "tok"}}).
		Resolve()
	if err := (Sink{Path: path}).Write(cat, resolved); err != nil {
		t.Fatal(err)
	}

	src, found, err := FromFile(path)
	if err != nil || !found {
		t.Fatalf("FromFile found=%v err=%v", found, err)
	}
	back := NewResolver(cat).With(Builtins(cat)).With(src).Resolve()
	if back.Address("hub").String() != "h:9" || !back.Toggle("tls") || back.Secret("token").Reveal() != "tok" {
		t.Fatalf("round-trip mismatch: hub=%q tls=%v token=%q",
			back.Address("hub").String(), back.Toggle("tls"), back.Secret("token").Reveal())
	}
}

func TestFromFileMissingIsNotAnError(t *testing.T) {
	_, found, err := FromFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil || found {
		t.Fatalf("missing file: found=%v err=%v", found, err)
	}
}

func TestWizardAnswersWinAndUseCurrentAsDefault(t *testing.T) {
	cat := sampleCatalog()
	current := NewResolver(cat).With(Builtins(cat)).Resolve()

	// First prompt (hub) gets empty input -> keeps the shown default; no other
	// option is askable, so the wizard stops there.
	prompter := NewPrompter(strings.NewReader("\n"), &strings.Builder{})
	wiz := RunWizard(cat, current, prompter)
	if v, ok := wiz.Value("hub"); !ok || v != "localhost:9090" {
		t.Fatalf("wizard hub = %q, %v (want kept default)", v, ok)
	}

	prompter = NewPrompter(strings.NewReader("typed-host:7\n"), &strings.Builder{})
	wiz = RunWizard(cat, current, prompter)
	r := NewResolver(cat).With(Builtins(cat)).With(wiz).Resolve()
	if got := r.Address("hub").String(); got != "typed-host:7" {
		t.Fatalf("typed wizard answer should win, got %q", got)
	}
}

func TestLocateHonorsConfigDirOverride(t *testing.T) {
	t.Setenv("HEIMDALL_CONFIG_DIR", "/tmp/heimdall-cfg")
	p, err := Locate("daemon")
	if err != nil {
		t.Fatal(err)
	}
	if p != "/tmp/heimdall-cfg/daemon.json" {
		t.Fatalf("Locate = %q", p)
	}
}

func TestProvidedIgnoresUnknownFlags(t *testing.T) {
	cat := sampleCatalog()
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	cat.Register(fs)
	fs.Bool("once", false, "action flag not in catalog")
	_ = fs.Parse([]string{"--once"})

	if Provided(cat, fs) {
		t.Fatal("an action flag must not count as a setting")
	}
	_ = fs.Parse([]string{"--hub", "x:1"})
	if !Provided(cat, fs) {
		t.Fatal("a catalog flag must count as provided")
	}
}
