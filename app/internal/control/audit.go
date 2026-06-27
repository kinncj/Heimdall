// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package control

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// AuditEntry records one control-plane invocation, allowed or refused.
type AuditEntry struct {
	Time     time.Time
	Actor    string
	Command  string
	Args     []string
	Allowed  bool
	ExitCode int
}

// Auditor records control-plane invocations for accountability. Every
// invocation passes through an Auditor before a result is returned.
type Auditor interface {
	Record(entry AuditEntry)
}

// WriterAuditor writes one line per invocation to an io.Writer (stderr, a file,
// or a syslog writer). It is safe for concurrent use.
type WriterAuditor struct {
	mu sync.Mutex
	w  io.Writer
}

// NewWriterAuditor returns an auditor that appends to w.
func NewWriterAuditor(w io.Writer) *WriterAuditor { return &WriterAuditor{w: w} }

func (a *WriterAuditor) Record(e AuditEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	decision := "ALLOW"
	if !e.Allowed {
		decision = "REFUSE"
	}
	fmt.Fprintf(a.w, "audit %s actor=%q cmd=%s args=[%s] decision=%s exit=%d\n",
		e.Time.UTC().Format(time.RFC3339), e.Actor, e.Command, strings.Join(e.Args, " "), decision, e.ExitCode)
}
