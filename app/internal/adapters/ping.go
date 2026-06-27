// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

// rePingRTT matches the round-trip time in a ping reply line, e.g. "time=4.513 ms"
// or "time<1 ms".
var rePingRTT = regexp.MustCompile(`time[=<]([0-9.]+)\s*ms`)

// pingArgs builds the platform-appropriate `ping` arguments for a single echo
// with the given per-probe timeout. Linux uses -W (reply wait), the BSDs and
// macOS use -t (total time).
func pingArgs(host string, timeout time.Duration) []string {
	sec := int(timeout.Seconds())
	if sec < 1 {
		sec = 1
	}
	switch runtime.GOOS {
	case "linux":
		return []string{"-c", "1", "-W", strconv.Itoa(sec), host}
	default: // darwin, *bsd
		return []string{"-c", "1", "-t", strconv.Itoa(sec), host}
	}
}

// parsePingRTT extracts the round-trip latency in milliseconds from ping output.
func parsePingRTT(out string) (float64, bool) {
	m := rePingRTT.FindStringSubmatch(out)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// pingFn performs one ICMP echo to host and returns the round-trip time in
// milliseconds. It is a package var so tests can substitute a deterministic
// implementation. ok is false when the host did not reply within timeout.
var pingFn = systemPing

func systemPing(ctx context.Context, host string, timeout time.Duration) (float64, bool) {
	cctx, cancel := context.WithTimeout(ctx, timeout+500*time.Millisecond)
	defer cancel()
	out, _ := exec.CommandContext(cctx, "ping", pingArgs(host, timeout)...).Output()
	return parsePingRTT(string(out))
}
