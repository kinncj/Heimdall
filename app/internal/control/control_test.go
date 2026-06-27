// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package control

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func TestAllowlistedCommandRunsAndAudits(t *testing.T) {
	var log bytes.Buffer
	e := &Executor{
		Allow: Allowlist{"echo.test": {Key: "echo.test", Argv: []string{"echo", "heimdall"}}},
		Audit: NewWriterAuditor(&log),
	}
	res := e.Execute(context.Background(), "echo.test", nil, "operator@dash")
	if res.Status != domain.StatusOK || res.ExitCode != 0 {
		t.Fatalf("result = %+v", res)
	}
	if strings.TrimSpace(res.Stdout) != "heimdall" {
		t.Errorf("stdout = %q, want heimdall", res.Stdout)
	}
	if !strings.Contains(log.String(), "decision=ALLOW") || !strings.Contains(log.String(), "operator@dash") {
		t.Errorf("audit missing allow entry: %q", log.String())
	}
}

func TestNonAllowlistedRefusedAndNotRun(t *testing.T) {
	var log bytes.Buffer
	e := &Executor{Allow: DefaultAllowlist(), Audit: NewWriterAuditor(&log)}
	res := e.Execute(context.Background(), "rm.rf", []string{"/"}, "mallory")
	if res.Status != domain.StatusInsufficientPermission {
		t.Fatalf("status = %v, want insufficient_permission", res.Status)
	}
	if res.Stdout != "" {
		t.Errorf("refused command produced output: %q", res.Stdout)
	}
	if !strings.Contains(log.String(), "decision=REFUSE") {
		t.Errorf("audit missing refuse entry: %q", log.String())
	}
}

func TestSudoIsNeverAllowlisted(t *testing.T) {
	a := DefaultAllowlist()
	if _, err := a.Resolve("sudo", []string{"reboot"}); err == nil {
		t.Fatal("sudo must never resolve")
	}
	for _, k := range a.Keys() {
		for _, tok := range a[k].Argv {
			if tok == "sudo" {
				t.Fatalf("allow-list command %q contains sudo", k)
			}
		}
	}
}

func TestDirListBoundedToAllowedRoots(t *testing.T) {
	a := DefaultAllowlist()
	if _, err := a.Resolve("dir.list", []string{"/root/.ssh"}); err == nil {
		t.Fatal("dir.list outside allowed roots must be refused")
	}
	if _, err := a.Resolve("dir.list", []string{"/var/log"}); err != nil {
		t.Fatalf("dir.list /var/log should be allowed: %v", err)
	}
	if _, err := a.Resolve("dir.list", []string{"/var/log/../../etc/shadow"}); err == nil {
		t.Fatal("path traversal escaping the roots must be refused")
	}
}

func TestNoArgsCommandRejectsArgs(t *testing.T) {
	a := DefaultAllowlist()
	if _, err := a.Resolve("uptime", []string{"; rm -rf /"}); err == nil {
		t.Fatal("uptime must reject extra args (no shell injection surface)")
	}
}
