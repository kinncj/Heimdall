// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"strings"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func cmdReg(t *testing.T, withCmd bool) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	labels := map[string]string{}
	if withCmd {
		labels["_cmd"] = "1"
	}
	reg.Enroll(domain.Host{ID: "h", DisplayName: "h", Context: domain.HostContext{Labels: labels}}, now)
	reg.Observe("h", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 1}}, nil, now)
	reg.Evaluate(now)
	return reg
}

func TestCommandModalRunsAndUnwinds(t *testing.T) {
	var gotHost, gotCmd, gotReq string
	m := Model{reg: cmdReg(t, true), detail: true, height: 40, width: 100, now: time.Unix(1, 5)}
	m.runCmd = func(host, cmd string, args []string, reqID string) {
		gotHost, gotCmd, gotReq = host, cmd, reqID
	}

	if m = press(m, "c"); m.modal != modalCmdList {
		t.Fatalf("c should open the command list, got %d", m.modal)
	}
	if m = press(m, "enter"); m.modal != modalCmdResult {
		t.Fatalf("enter should run and show the result, got %d", m.modal)
	}
	if gotHost != "h" || gotCmd != "process.list" || gotReq == "" {
		t.Fatalf("runCmd called with host=%q cmd=%q req=%q", gotHost, gotCmd, gotReq)
	}

	// Before the result arrives, the modal says "running".
	h, _ := m.selectedHost()
	if body := strings.Join(m.cmdResultBody(h, 100), "\n"); !strings.Contains(body, "running") {
		t.Fatalf("expected running state, got: %s", body)
	}
	// When the matching result lands in the registry, it renders.
	m.reg.RecordCommandResult("h", &domain.CommandResult{RequestID: gotReq, Status: domain.StatusOK, Stdout: "hello-output"})
	h, _ = m.selectedHost()
	if body := strings.Join(m.cmdResultBody(h, 100), "\n"); !strings.Contains(body, "hello-output") {
		t.Fatalf("result not rendered: %s", body)
	}

	// esc unwinds: result -> list -> detail.
	if m = press(m, "esc"); m.modal != modalCmdList {
		t.Fatalf("esc from result should return to the list, got %d", m.modal)
	}
	if m = press(m, "esc"); m.modal != modalNone {
		t.Fatalf("esc from the list should close the modal, got %d", m.modal)
	}
}

func TestCommandModalGatedByCapabilityAndCallback(t *testing.T) {
	// No _cmd label: c does nothing even with a callback.
	m := Model{reg: cmdReg(t, false), detail: true, height: 40, width: 100}
	m.runCmd = func(string, string, []string, string) {}
	if m = press(m, "c"); m.modal != modalNone {
		t.Fatal("c must do nothing without the _cmd capability")
	}
	// _cmd present but no callback (e.g. no hub): also gated.
	m2 := Model{reg: cmdReg(t, true), detail: true, height: 40, width: 100}
	if m2 = press(m2, "c"); m2.modal != modalNone {
		t.Fatal("c must do nothing without a run callback")
	}
}
