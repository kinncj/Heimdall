// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import "time"

// ProcessRow is one row of a host's pushed process table (Heimdallr's sight,
// ADR 0017): pid/ppid, the process's CPU% and memory%, and the command. Hosts
// push it periodically; nothing requests it.
type ProcessRow struct {
	PID     uint32
	PPID    uint32
	CPUPct  float64
	MemPct  float64
	Command string
}

// LogLine is one tailed log line pushed from a host. At is the daemon-side
// timestamp; RateLimited marks that lines were dropped before this one.
type LogLine struct {
	Source      string
	At          time.Time
	Line        string
	Level       string
	RateLimited bool
}
