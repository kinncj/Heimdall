// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"heimdall/app/internal/command"
	"heimdall/app/internal/domain"
)

func startHelper(t *testing.T) string {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "h.sock")
	srv := &Server{SockPath: sock, Collect: func(context.Context) []domain.Metric { return nil }}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Serve(ctx) }()
	// wait for the socket to come up
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := (Client{SockPath: sock}).Collect(context.Background()); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	return sock
}

// The helper trusts no caller: it refuses to run a command that is not on its own
// privileged allow-list, even though the daemon asked (v2 Phase 2b).
func TestExecRefusesNonPrivileged(t *testing.T) {
	sock := startHelper(t)
	res, err := (Client{SockPath: sock}).Exec(context.Background(), "process.list", nil)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if res.Status != domain.StatusInsufficientPermission {
		t.Fatalf("status = %v, want insufficient_permission", res.Status)
	}
	if !strings.Contains(res.Stderr, "not a privileged") {
		t.Fatalf("stderr = %q, want a not-privileged refusal", res.Stderr)
	}
}

// A privileged command is routed to execution (not refused on the allow-list).
// Whether it then succeeds depends on the host, so we only assert it was attempted.
func TestExecRunsPrivilegedCommand(t *testing.T) {
	sock := startHelper(t)
	res, err := (Client{SockPath: sock}).Exec(context.Background(), "dmesg", nil)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if strings.Contains(res.Stderr, "not a privileged allow-listed") {
		t.Fatalf("a privileged command must not be refused by the allow-list: %q", res.Stderr)
	}
}

func TestExecUnavailableWhenAbsent(t *testing.T) {
	_, err := (Client{SockPath: filepath.Join(t.TempDir(), "missing.sock")}).Exec(context.Background(), "dmesg", nil)
	if err != ErrUnavailable {
		t.Fatalf("got %v, want ErrUnavailable", err)
	}
}

func TestIsPrivileged(t *testing.T) {
	for _, k := range []string{"dmesg", "journal.tail"} {
		if !command.IsPrivileged(k) {
			t.Errorf("%s should be privileged", k)
		}
	}
	for _, k := range []string{"process.list", "disk.df", "uptime", "os.info"} {
		if command.IsPrivileged(k) {
			t.Errorf("%s should not be privileged", k)
		}
	}
}
