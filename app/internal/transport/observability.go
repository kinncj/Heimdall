// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package transport

import (
	"time"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

// ProcessRowToProto maps a domain process row to the wire type.
func ProcessRowToProto(p domain.ProcessRow) *v1.ProcessRow {
	return &v1.ProcessRow{
		Pid: p.PID, Ppid: p.PPID,
		CpuPct: float32(p.CPUPct), MemPct: float32(p.MemPct),
		Command: p.Command,
	}
}

// ProcessRowFromProto maps a wire process row back to the domain type.
func ProcessRowFromProto(p *v1.ProcessRow) domain.ProcessRow {
	return domain.ProcessRow{
		PID: p.GetPid(), PPID: p.GetPpid(),
		CPUPct: float64(p.GetCpuPct()), MemPct: float64(p.GetMemPct()),
		Command: p.GetCommand(),
	}
}

// LogLineToProto maps a domain log line to the wire type, stamping the host id.
func LogLineToProto(hostID string, l domain.LogLine) *v1.LogLine {
	return &v1.LogLine{
		HostId: hostID, Source: l.Source,
		TsUnixMillis: l.At.UnixMilli(), Line: l.Line, Level: l.Level,
		RateLimited: l.RateLimited,
	}
}

// LogLineFromProto maps a wire log line back to the domain type.
func LogLineFromProto(l *v1.LogLine) domain.LogLine {
	return domain.LogLine{
		Source: l.GetSource(), At: time.UnixMilli(l.GetTsUnixMillis()),
		Line: l.GetLine(), Level: l.GetLevel(), RateLimited: l.GetRateLimited(),
	}
}

// AttachObservability sets the pushed process table and log lines on a snapshot
// (Heimdallr's sight). Empty inputs leave the snapshot untouched, so the fields
// stay absent on the wire for daemons that push nothing.
func AttachObservability(s *v1.Snapshot, procs []domain.ProcessRow, procsAt time.Time, hostID string, logs []domain.LogLine) {
	if len(procs) > 0 {
		rows := make([]*v1.ProcessRow, 0, len(procs))
		for _, p := range procs {
			rows = append(rows, ProcessRowToProto(p))
		}
		s.Processes = rows
		s.ProcessesAtUnixMillis = procsAt.UnixMilli()
	}
	if len(logs) > 0 {
		lines := make([]*v1.LogLine, 0, len(logs))
		for _, l := range logs {
			lines = append(lines, LogLineToProto(hostID, l))
		}
		s.LogLines = lines
	}
}

// CommandResultFromSnapshot extracts an on-demand command result (v2 Phase 2), or
// nil when the snapshot carries none.
func CommandResultFromSnapshot(s *v1.Snapshot) *domain.CommandResult {
	cr := s.GetCommandResult()
	if cr == nil {
		return nil
	}
	return &domain.CommandResult{
		RequestID: cr.GetRequestId(),
		ExitCode:  int(cr.GetExitCode()),
		Stdout:    cr.GetStdout(),
		Stderr:    cr.GetStderr(),
		Truncated: cr.GetTruncated(),
		Status:    statusFromProto(cr.GetStatus()),
	}
}

// ObservabilityFromSnapshot extracts the pushed process table (with its collection
// time) and log lines from a wire snapshot.
func ObservabilityFromSnapshot(s *v1.Snapshot) ([]domain.ProcessRow, time.Time, []domain.LogLine) {
	var procs []domain.ProcessRow
	if rows := s.GetProcesses(); len(rows) > 0 {
		procs = make([]domain.ProcessRow, 0, len(rows))
		for _, p := range rows {
			procs = append(procs, ProcessRowFromProto(p))
		}
	}
	var logs []domain.LogLine
	if lines := s.GetLogLines(); len(lines) > 0 {
		logs = make([]domain.LogLine, 0, len(lines))
		for _, l := range lines {
			logs = append(logs, LogLineFromProto(l))
		}
	}
	return procs, time.UnixMilli(s.GetProcessesAtUnixMillis()), logs
}
