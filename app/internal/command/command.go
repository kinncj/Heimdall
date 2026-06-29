// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package command runs the on-demand, READ-ONLY, allow-listed commands of the v2
// socket transport (ADR 0018). A logical key maps to a fixed, OS-appropriate
// argument vector. A few Windows entries shell out via "cmd /c", but always with
// a constant argument list — no caller input is ever interpolated into a command
// line, so there is no shell-injection surface. Commands run as the daemon's
// unprivileged user; output is bounded. Nothing arbitrary is ever executed.
package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"heimdall/app/internal/domain"
)

const (
	defaultTimeout   = 8 * time.Second
	defaultOutputCap = 64 * 1024
)

// Result is the bounded outcome of a command.
type Result struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	Truncated bool
	Status    domain.MetricStatus
}

// spec is one allow-listed command: a per-OS argv builder, optional argument
// validation, and whether it needs privilege. A nil argv for the current OS means
// "unavailable here". Privileged commands are run by the root helper, never by the
// unprivileged daemon (v2 Phase 2b).
type spec struct {
	argv       func(goos string) []string
	validate   func(args []string) ([]string, error)
	privileged bool
}

// Keys lists the allow-listed command keys (stable order) for discovery/help.
func Keys() []string {
	a := allowlist()
	order := []string{"process.list", "disk.df", "uptime", "os.info", "dir.list", "dmesg", "journal.tail"}
	out := make([]string, 0, len(order))
	for _, k := range order {
		if _, ok := a[k]; ok {
			out = append(out, k)
		}
	}
	return out
}

// IsPrivileged reports whether a command must run via the privileged helper.
func IsPrivileged(key string) bool { return allowlist()[key].privileged }

func allowlist() map[string]spec {
	return map[string]spec{
		"process.list": {argv: func(os string) []string {
			if os == "windows" {
				return []string{"tasklist"}
			}
			return []string{"ps", "-eo", "pid,ppid,pcpu,pmem,comm"}
		}},
		"disk.df": {argv: func(os string) []string {
			if os == "windows" {
				return []string{"wmic", "logicaldisk", "get", "caption,freespace,size"}
			}
			return []string{"df", "-h"}
		}},
		"uptime": {argv: func(os string) []string {
			if os == "windows" {
				return nil
			}
			return []string{"uptime"}
		}},
		"os.info": {argv: func(os string) []string {
			if os == "windows" {
				return []string{"cmd", "/c", "ver"}
			}
			return []string{"uname", "-a"}
		}},
		"dir.list": {
			argv: func(os string) []string {
				if os == "windows" {
					return []string{"cmd", "/c", "dir"}
				}
				return []string{"ls", "-la"}
			},
			validate: allowedDirArg,
		},
		// Privileged (run via the root helper, v2 Phase 2b) — Linux-only kernel/log
		// reads that typically need root.
		"dmesg": {argv: func(os string) []string {
			if os != "linux" {
				return nil
			}
			return []string{"dmesg", "--ctime"}
		}, privileged: true},
		"journal.tail": {argv: func(os string) []string {
			if os != "linux" {
				return nil
			}
			return []string{"journalctl", "-n", "200", "--no-pager"}
		}, privileged: true},
	}
}

// Resolve validates a key+args and returns the full argv, or an error. Exposed so
// the allow-list can be tested without executing anything.
func Resolve(goos, key string, args []string) ([]string, error) {
	s, ok := allowlist()[key]
	if !ok {
		return nil, fmt.Errorf("command %q is not allow-listed", key)
	}
	base := s.argv(goos)
	if base == nil {
		return nil, fmt.Errorf("command %q is not available on %s", key, goos)
	}
	argv := append([]string(nil), base...)
	if s.validate == nil {
		if len(args) > 0 {
			return nil, fmt.Errorf("command %q accepts no arguments", key)
		}
		return argv, nil
	}
	tail, err := s.validate(args)
	if err != nil {
		return nil, err
	}
	return append(argv, tail...), nil
}

// Run executes an allow-listed command and returns a bounded result. A rejected
// command (unknown key, bad args, unavailable on this OS) yields
// INSUFFICIENT_PERMISSION and is never executed.
func Run(ctx context.Context, key string, args []string) Result {
	argv, err := Resolve(runtime.GOOS, key, args)
	if err != nil {
		return Result{Status: domain.StatusInsufficientPermission, Stderr: err.Error(), ExitCode: -1}
	}
	cctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, argv[0], argv[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()

	exit := 0
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			exit = ee.ExitCode()
		} else {
			return Result{Status: domain.StatusError, Stderr: runErr.Error(), ExitCode: -1}
		}
	}
	out, t1 := capString(stdout.String())
	errOut, t2 := capString(stderr.String())
	return Result{
		ExitCode:  exit,
		Stdout:    out,
		Stderr:    errOut,
		Truncated: t1 || t2,
		Status:    domain.StatusOK,
	}
}

func capString(s string) (string, bool) {
	if len(s) <= defaultOutputCap {
		return s, false
	}
	return s[:defaultOutputCap], true
}

// allowedDirRoots bounds dir.list to non-sensitive, operator-relevant trees.
var allowedDirRoots = []string{"/var/log", "/tmp"}

// allowedDirArg permits exactly one absolute path within an allow-listed root,
// cleaned first so traversal that escapes the roots is refused.
func allowedDirArg(args []string) ([]string, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dir.list takes exactly one directory")
	}
	clean := filepath.Clean(args[0])
	if !filepath.IsAbs(clean) {
		return nil, fmt.Errorf("directory must be an absolute path")
	}
	for _, root := range allowedDirRoots {
		if clean == root || strings.HasPrefix(clean, root+string(filepath.Separator)) {
			return []string{clean}, nil
		}
	}
	return nil, fmt.Errorf("directory %q is not within an allow-listed root", clean)
}
