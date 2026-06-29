// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package command

import (
	"context"
	"testing"

	"heimdall/app/internal/domain"
)

func TestResolveAllowlist(t *testing.T) {
	if _, err := Resolve("linux", "process.list", nil); err != nil {
		t.Fatalf("process.list should resolve: %v", err)
	}
	if _, err := Resolve("linux", "rm.rf", nil); err == nil {
		t.Fatal("an unknown command must be rejected")
	}
	if _, err := Resolve("linux", "process.list", []string{"extra"}); err == nil {
		t.Fatal("args to a no-arg command must be rejected")
	}
	if _, err := Resolve("windows", "uptime", nil); err == nil {
		t.Fatal("uptime is unavailable on windows and must be rejected")
	}
}

func TestDirListIsBounded(t *testing.T) {
	if _, err := Resolve("linux", "dir.list", []string{"/var/log"}); err != nil {
		t.Fatalf("/var/log should be allowed: %v", err)
	}
	if _, err := Resolve("linux", "dir.list", []string{"/etc/shadow"}); err == nil {
		t.Fatal("a path outside the allow-listed roots must be rejected")
	}
	if _, err := Resolve("linux", "dir.list", []string{"/var/log/../../etc"}); err == nil {
		t.Fatal("traversal escaping the roots must be rejected")
	}
	if _, err := Resolve("linux", "dir.list", []string{"relative"}); err == nil {
		t.Fatal("a relative path must be rejected")
	}
}

func TestRunRejectsUnknown(t *testing.T) {
	res := Run(context.Background(), "definitely.not.allowed", nil)
	if res.Status != domain.StatusInsufficientPermission {
		t.Fatalf("status = %v, want insufficient_permission", res.Status)
	}
	if res.ExitCode != -1 {
		t.Fatalf("rejected command should not run (exit -1), got %d", res.ExitCode)
	}
}

func TestRunProcessList(t *testing.T) {
	// On the test host (unix/windows CI) this actually runs ps/tasklist.
	res := Run(context.Background(), "process.list", nil)
	if res.Status != domain.StatusOK {
		t.Fatalf("process.list status = %v (stderr: %s)", res.Status, res.Stderr)
	}
	if res.Stdout == "" {
		t.Fatal("process.list should produce output")
	}
}

func TestKeys(t *testing.T) {
	keys := Keys()
	if len(keys) == 0 || keys[0] != "process.list" {
		t.Fatalf("keys = %v", keys)
	}
}
