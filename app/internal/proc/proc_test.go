// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package proc

import "testing"

func TestParsePS(t *testing.T) {
	out := []byte(`  PID  PPID %CPU %MEM COMMAND
    1     0  0.0  0.1 systemd
  421     1 12.5  3.4 heimdall daemon
`)
	rows := Parse("linux", out)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	if rows[1].PID != 421 || rows[1].PPID != 1 {
		t.Fatalf("pid/ppid = %d/%d", rows[1].PID, rows[1].PPID)
	}
	if rows[1].CPUPct != 12.5 || rows[1].MemPct != 3.4 {
		t.Fatalf("cpu/mem = %v/%v", rows[1].CPUPct, rows[1].MemPct)
	}
	if rows[1].Command != "heimdall daemon" { // command with a space is preserved
		t.Fatalf("command = %q", rows[1].Command)
	}
}

func TestParseTasklist(t *testing.T) {
	out := []byte("\"System Idle Process\",\"0\",\"Services\",\"0\",\"8 K\"\r\n" +
		"\"heimdall-daemon.exe\",\"1234\",\"Console\",\"1\",\"45,120 K\"\r\n")
	rows := Parse("windows", out)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	if rows[1].PID != 1234 || rows[1].Command != "heimdall-daemon.exe" {
		t.Fatalf("row = %+v", rows[1])
	}
}

func TestArgvFor(t *testing.T) {
	if argvFor("windows")[0] != "tasklist" {
		t.Error("windows fallback should use tasklist")
	}
	for _, os := range []string{"linux", "darwin"} {
		if argvFor(os)[0] != "ps" {
			t.Errorf("%s should use ps", os)
		}
	}
}

func TestParseWindowsPerf(t *testing.T) {
	// Lines emitted by the perf-counter PowerShell query: pid,cpu%,mem%,name.
	out := []byte("4,0,0.0,System\r\n" +
		"1234,12,3.4,heimdall-daemon\r\n" +
		"5678,1,0.2,svchost#3\r\n")
	rows := parseWindowsPerf(out)
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3: %+v", len(rows), rows)
	}
	r := rows[1]
	if r.PID != 1234 || r.CPUPct != 12 || r.MemPct != 3.4 || r.Command != "heimdall-daemon" {
		t.Fatalf("row = %+v", r)
	}
}

func TestParseWindowsPerf_SkipsGarbageAndBlank(t *testing.T) {
	out := []byte("\r\nnot,enough\r\n9,5,1.0,proc\r\n")
	rows := parseWindowsPerf(out)
	if len(rows) != 1 || rows[0].PID != 9 || rows[0].CPUPct != 5 {
		t.Fatalf("rows = %+v", rows)
	}
}
