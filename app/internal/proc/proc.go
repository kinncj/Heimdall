// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package proc collects a host's process table with an OS-appropriate command,
// for the dashboard's top modal (Heimdallr's sight, ADR 0017). It is read-only
// and unprivileged; a privileged helper-backed Source can enrich it later.
package proc

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"heimdall/app/internal/domain"
)

// Source yields a host's process table. The default Local source shells out to an
// OS command; a future helper-backed source (privileged) can satisfy the same
// contract to show fuller detail on hosts that run heimdall-helper (DIP).
type Source interface {
	Collect(ctx context.Context) ([]domain.ProcessRow, error)
}

// Local collects the process table on the daemon's own host.
type Local struct{}

// Collect runs the OS-appropriate process command and parses its output. On
// Windows it first asks the performance-counter class for real CPU% and memory
// (tasklist exposes neither), falling back to tasklist for pid+name only when
// that query is unavailable.
func (Local) Collect(ctx context.Context) ([]domain.ProcessRow, error) {
	if runtime.GOOS == "windows" {
		argv := winPerfArgv()
		if out, err := exec.CommandContext(ctx, argv[0], argv[1:]...).Output(); err == nil {
			if rows := parseWindowsPerf(out); len(rows) > 0 {
				return rows, nil
			}
		}
		// fall through to tasklist (pid + name only) when the perf query fails.
	}
	argv := argvFor(runtime.GOOS)
	out, err := exec.CommandContext(ctx, argv[0], argv[1:]...).Output()
	if err != nil {
		return nil, err
	}
	return Parse(runtime.GOOS, out), nil
}

// winPerfArgv builds the PowerShell query for per-process CPU% and memory%.
// PercentProcessorTime is instantaneous (0–100×cores, like ps %CPU); WorkingSet
// is divided by total physical memory for a percentage that matches the Unix
// pmem column. The _Total/Idle rollups and pid 0 are dropped.
func winPerfArgv() []string {
	const ps = `$m=(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory; ` +
		`Get-CimInstance Win32_PerfFormattedData_PerfProc_Process | ` +
		`Where-Object { $_.IDProcess -ne 0 -and $_.Name -ne '_Total' } | ` +
		`ForEach-Object { '{0},{1},{2},{3}' -f $_.IDProcess,$_.PercentProcessorTime,` +
		`[math]::Round($_.WorkingSet/$m*100,1),$_.Name }`
	return []string{"powershell", "-NoProfile", "-Command", ps}
}

// parseWindowsPerf reads the perf-counter query's "pid,cpu,mem,name" lines. The
// name is taken as the remainder so a process whose name contains a comma is
// still kept whole.
func parseWindowsPerf(out []byte) []domain.ProcessRow {
	lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	rows := make([]domain.ProcessRow, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimRight(ln, "\r")
		if strings.TrimSpace(ln) == "" {
			continue
		}
		f := strings.SplitN(ln, ",", 4)
		if len(f) < 4 {
			continue
		}
		rows = append(rows, domain.ProcessRow{
			PID:     atou(f[0]),
			CPUPct:  atof(f[1]),
			MemPct:  atof(f[2]),
			Command: strings.TrimSpace(f[3]),
		})
	}
	return rows
}

// argvFor returns the per-OS process-listing command. Unix uses ps with a fixed
// column set; Windows uses tasklist in headerless CSV as the pid+name fallback
// (the primary Windows path is winPerfArgv). Selection is a lookup, not a
// call-site branch.
func argvFor(goos string) []string {
	switch goos {
	case "windows":
		return []string{"tasklist", "/FO", "CSV", "/NH"}
	default: // linux, darwin, *bsd
		return []string{"ps", "-eo", "pid,ppid,pcpu,pmem,comm"}
	}
}

// Parse turns process-command output into rows. It is total: unparseable lines
// are skipped rather than failing the whole collection.
func Parse(goos string, out []byte) []domain.ProcessRow {
	if goos == "windows" {
		return parseTasklist(out)
	}
	return parsePS(out)
}

// parsePS reads `ps -eo pid,ppid,pcpu,pmem,comm` output: a header line then rows
// of five whitespace-separated columns (the command may itself contain spaces).
func parsePS(out []byte) []domain.ProcessRow {
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	rows := make([]domain.ProcessRow, 0, len(lines))
	for i, ln := range lines {
		if i == 0 || strings.TrimSpace(ln) == "" {
			continue // header / blank
		}
		f := strings.Fields(ln)
		if len(f) < 5 {
			continue
		}
		rows = append(rows, domain.ProcessRow{
			PID:     atou(f[0]),
			PPID:    atou(f[1]),
			CPUPct:  atof(f[2]),
			MemPct:  atof(f[3]),
			Command: strings.Join(f[4:], " "),
		})
	}
	return rows
}

// parseTasklist reads `tasklist /FO CSV /NH` rows:
// "image","pid","session","sess#","mem usage". CPU% and ppid are unavailable.
func parseTasklist(out []byte) []domain.ProcessRow {
	lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	rows := make([]domain.ProcessRow, 0, len(lines))
	for _, ln := range lines {
		cols := splitCSV(strings.TrimRight(ln, "\r"))
		if len(cols) < 2 {
			continue
		}
		rows = append(rows, domain.ProcessRow{
			PID:     atou(cols[1]),
			Command: cols[0],
		})
	}
	return rows
}

// splitCSV splits a simple quoted CSV line (no embedded quotes), as tasklist emits.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.Trim(p, `"`)
	}
	return parts
}

func atou(s string) uint32 {
	n, _ := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	return uint32(n)
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}
