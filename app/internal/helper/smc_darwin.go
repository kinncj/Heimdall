// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build darwin && cgo

package helper

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <IOKit/IOKitLib.h>
#include <string.h>
#include <stdint.h>

// AppleSMC key access. PSTR ("System Total Power", IEEE float, watts) is the
// whole-system power rail mactop/btop/iStat read without root — and the only
// power figure populated on Apple Silicon Pro/Max chips, whose IOReport energy
// channels report 0 for CPU/ANE.
typedef struct { char major, minor, build, reserved[1]; uint16_t release; } smc_vers_t;
typedef struct { uint16_t version, length; uint32_t cpuPLimit, gpuPLimit, memPLimit; } smc_plim_t;
typedef struct { uint32_t dataSize, dataType; char dataAttributes; } smc_kinfo_t;
typedef struct {
	uint32_t key;
	smc_vers_t vers;
	smc_plim_t plim;
	smc_kinfo_t keyInfo;
	char result, status, data8;
	uint32_t data32;
	uint8_t bytes[32];
} smc_kd_t;

// AppleSMC user-client selector and command codes.
enum { kSMCHandleYPCEvent = 2, kSMCReadKeyInfo = 9, kSMCReadBytes = 5 };

static uint32_t smc_str2key(const char *s) {
	return ((uint32_t)s[0] << 24) | ((uint32_t)s[1] << 16) | ((uint32_t)s[2] << 8) | (uint32_t)s[3];
}

// smc_read_pstr opens AppleSMC, reads PSTR as a float, and returns watts. The
// connection is opened and closed per call; this runs at the collector cadence
// (hundreds of ms), so the open cost is negligible and we avoid holding a port.
// Returns 0 on any failure (no AppleSMC, key absent, wrong type).
static double smc_read_pstr(void) {
	io_service_t svc = IOServiceGetMatchingService(0, IOServiceMatching("AppleSMC"));
	if (!svc) return 0;
	io_connect_t conn = 0;
	if (IOServiceOpen(svc, mach_task_self(), 0, &conn) != KERN_SUCCESS) {
		IOObjectRelease(svc);
		return 0;
	}
	IOObjectRelease(svc);

	uint32_t key = smc_str2key("PSTR");
	double watts = 0;

	smc_kd_t in = {0}, out = {0};
	in.key = key;
	in.data8 = kSMCReadKeyInfo;
	size_t os = sizeof(smc_kd_t);
	if (IOConnectCallStructMethod(conn, kSMCHandleYPCEvent, &in, sizeof(smc_kd_t), &out, &os) == KERN_SUCCESS) {
		char t[5] = {0};
		t[0] = out.keyInfo.dataType >> 24;
		t[1] = out.keyInfo.dataType >> 16;
		t[2] = out.keyInfo.dataType >> 8;
		t[3] = out.keyInfo.dataType;
		if (out.keyInfo.dataSize == 4 && !strcmp(t, "flt ")) {
			smc_kd_t ri = {0}, ro = {0};
			ri.key = key;
			ri.keyInfo.dataSize = out.keyInfo.dataSize;
			ri.keyInfo.dataType = out.keyInfo.dataType;
			ri.data8 = kSMCReadBytes;
			os = sizeof(smc_kd_t);
			if (IOConnectCallStructMethod(conn, kSMCHandleYPCEvent, &ri, sizeof(smc_kd_t), &ro, &os) == KERN_SUCCESS) {
				float f;
				memcpy(&f, ro.bytes, 4);
				watts = (double)f;
			}
		}
	}
	IOServiceClose(conn);
	return watts;
}
*/
import "C"

// smcSystemPower reads the SMC PSTR ("System Total Power") rail in watts. No root
// required. ok is false when AppleSMC or the key is unavailable, or the reading
// is non-positive, so callers fall back to IOReport / powermetrics.
func smcSystemPower() (watts float64, ok bool) {
	w := float64(C.smc_read_pstr())
	if w <= 0 {
		return 0, false
	}
	return w, true
}
