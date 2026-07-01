// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import "testing"

// Scaphandre's Prometheus exporter reports per-socket CPU power in microwatts;
// power.cpu is their sum in watts.
func TestParseScaphandreCPUWatts(t *testing.T) {
	text := `# HELP scaph_socket_power_microwatts Power measured per CPU socket
# TYPE scaph_socket_power_microwatts gauge
scaph_host_power_microwatts 45000000
scaph_socket_power_microwatts{socket_id="0"} 30000000
scaph_socket_power_microwatts{socket_id="1"} 10000000
`
	w, ok := parseScaphandreCPUWatts(text)
	if !ok || w != 40 { // (30M + 10M) µW ÷ 1e6 = 40 W
		t.Fatalf("got %v %v, want 40 W", w, ok)
	}

	// Scientific notation (Scaphandre emits floats) is accepted.
	w, ok = parseScaphandreCPUWatts("scaph_socket_power_microwatts{socket_id=\"0\"} 1.25e7\n")
	if !ok || w != 12.5 {
		t.Fatalf("got %v %v, want 12.5 W", w, ok)
	}

	// No socket metric present → no reading (not a phantom 0).
	if _, ok := parseScaphandreCPUWatts("# nothing here\nscaph_host_power_microwatts 1000000\n"); ok {
		t.Error("no socket power metric should yield no reading")
	}
}
