// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"strconv"
	"strings"
)

// parseScaphandreCPUWatts sums Scaphandre's per-socket CPU power from its
// Prometheus exposition and returns watts. Scaphandre reports
// scaph_socket_power_microwatts per CPU socket (the RAPL package domain) in
// microwatts; summing the sockets and dividing by 1e6 gives CPU-package power —
// the Windows equivalent of the RAPL package read directly on Linux. No socket
// line means no reading.
func parseScaphandreCPUWatts(promText string) (float64, bool) {
	var micro float64
	found := false
	for _, ln := range strings.Split(promText, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		if !strings.HasPrefix(ln, "scaph_socket_power_microwatts") {
			continue
		}
		f := strings.Fields(ln)
		if len(f) < 2 {
			continue
		}
		v, err := strconv.ParseFloat(f[len(f)-1], 64)
		if err != nil {
			continue
		}
		micro += v
		found = true
	}
	if !found {
		return 0, false
	}
	return micro / 1e6, true
}
