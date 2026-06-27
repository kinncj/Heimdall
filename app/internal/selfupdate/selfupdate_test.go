// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package selfupdate

import (
	"errors"
	"fmt"
	"io/fs"
	"runtime"
	"strings"
	"testing"
)

func TestNeedsElevationDetectsPermission(t *testing.T) {
	if !needsElevation(fmt.Errorf("cannot write: %w", fs.ErrPermission)) {
		t.Error("a wrapped permission error should need elevation")
	}
	if needsElevation(errors.New("network down")) {
		t.Error("a non-permission error should not need elevation")
	}
}

func TestElevationCommandUsesSudoOnUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}
	cmd, err := elevationCommand()
	if err != nil {
		t.Skipf("sudo not available in this environment: %v", err)
	}
	if !strings.Contains(cmd.Path, "sudo") {
		t.Errorf("elevation should go through sudo, got %q", cmd.Path)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "update") {
		t.Errorf("elevation should re-run the update subcommand, got %v", cmd.Args)
	}
}

func TestAssetNameMatchesReleaseNaming(t *testing.T) {
	got := assetName("daemon")
	want := "heimdall-daemon_" + runtime.GOOS + "_" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		want += ".exe"
	}
	if got != want {
		t.Fatalf("assetName = %q, want %q", got, want)
	}
}

func TestSumForParsesManifest(t *testing.T) {
	manifest := "abc123  heimdall-daemon_linux_amd64\n" +
		"def456  heimdall-hub_linux_amd64\n"

	if s, ok := sumFor(manifest, "heimdall-hub_linux_amd64"); !ok || s != "def456" {
		t.Fatalf("sumFor hub = %q, %v", s, ok)
	}
	if _, ok := sumFor(manifest, "heimdall-missing_linux_amd64"); ok {
		t.Fatal("sumFor should report missing asset as not found")
	}
}

func TestRepoHonorsOverride(t *testing.T) {
	t.Setenv("HEIMDALL_REPO", "acme/widget")
	if repo() != "acme/widget" {
		t.Fatalf("repo() = %q, want override", repo())
	}
}
