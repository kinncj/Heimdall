// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package control

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"heimdall/app/internal/domain"
)

const (
	defaultTimeout   = 5 * time.Second
	defaultOutputCap = 64 * 1024
)

// Result is the bounded outcome of a control invocation.
type Result struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	Truncated bool
	Status    domain.MetricStatus
}

// Executor validates control requests against an allow-list, runs allowed
// commands as the current (unprivileged) user, bounds their output, and audits
// every invocation — allowed or refused. It never invokes sudo and never uses a
// shell: argv[0] is executed directly with its fixed and validated arguments.
type Executor struct {
	Allow     Allowlist
	Audit     Auditor
	Timeout   time.Duration
	OutputCap int
}

// NewExecutor builds an executor with the default allow-list and the given
// auditor.
func NewExecutor(audit Auditor) *Executor {
	return &Executor{Allow: DefaultAllowlist(), Audit: audit, Timeout: defaultTimeout, OutputCap: defaultOutputCap}
}

// Execute resolves and runs the request. A refused request returns
// InsufficientPermission and is never executed.
func (e *Executor) Execute(ctx context.Context, key string, args []string, actor string) Result {
	argv, err := e.Allow.Resolve(key, args)
	if err != nil {
		e.record(actor, key, args, false, -1)
		return Result{Status: domain.StatusInsufficientPermission, Stderr: err.Error(), ExitCode: -1}
	}

	to := e.Timeout
	if to <= 0 {
		to = defaultTimeout
	}
	cctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	cmd := exec.CommandContext(cctx, argv[0], argv[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	exit := 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			e.record(actor, key, args, true, -1)
			return Result{Status: domain.StatusError, Stderr: runErr.Error(), ExitCode: -1}
		}
	}

	limit := e.OutputCap
	if limit <= 0 {
		limit = defaultOutputCap
	}
	out, trunc := capString(stdout.String(), limit)
	errOut, errTrunc := capString(stderr.String(), limit)
	e.record(actor, key, args, true, exit)
	return Result{ExitCode: exit, Stdout: out, Stderr: errOut, Truncated: trunc || errTrunc, Status: domain.StatusOK}
}

func (e *Executor) record(actor, key string, args []string, allowed bool, exit int) {
	if e.Audit == nil {
		return
	}
	e.Audit.Record(AuditEntry{
		Time: time.Now(), Actor: actor, Command: key, Args: append([]string(nil), args...),
		Allowed: allowed, ExitCode: exit,
	})
}

func capString(s string, n int) (string, bool) {
	if len(s) <= n {
		return s, false
	}
	return s[:n], true
}
