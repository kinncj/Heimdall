// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build darwin && cgo

package helper

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <IOKit/IOKitLib.h>
#include <string.h>
#include <stdint.h>
#include <stdlib.h>

// AppleSMC key access. PSTR ("System Total Power", IEEE float, watts) is the
// whole-system power rail mactop/btop/iStat read without root. The per-domain
// keys (PCPC "CPU Package", PCTR "CPU Total", PG0C "GPU", …) carry CPU/GPU power
// even on Apple Silicon Pro/Max, whose IOReport energy channels report 0 — the
// same raw SMC keys Stats reads. All are IEEE floats in watts.
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

// smc_read_key opens AppleSMC, reads the given 4-char key as a float, and returns
// watts. *found is set to 1 when the key exists and was read as a float (the
// value may legitimately be 0), else 0 — so callers can tell an absent key from a
// genuine zero. The connection is opened/closed per call; at the collector
// cadence the open cost is negligible and we avoid holding a port.
static double smc_read_key(const char *k, int *found) {
	*found = 0;
	io_service_t svc = IOServiceGetMatchingService(0, IOServiceMatching("AppleSMC"));
	if (!svc) return 0;
	io_connect_t conn = 0;
	if (IOServiceOpen(svc, mach_task_self(), 0, &conn) != KERN_SUCCESS) {
		IOObjectRelease(svc);
		return 0;
	}
	IOObjectRelease(svc);

	uint32_t key = smc_str2key(k);
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
				*found = 1;
			}
		}
	}
	IOServiceClose(conn);
	return watts;
}
*/
import "C"

import (
	"math"
	"unsafe"
)

// smcReadFloat reads a 4-char AppleSMC float key in watts. ok is true when the
// key exists and read as a float (value may be 0); false when absent or the wrong
// type. No root required.
func smcReadFloat(key string) (float64, bool) {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	var found C.int
	w := float64(C.smc_read_key(ck, &found))
	return w, found != 0
}

// smcSystemPower reads the SMC PSTR ("System Total Power") rail in watts. ok is
// false when AppleSMC or the key is unavailable, or the reading is non-positive,
// so callers fall back to IOReport / powermetrics.
func smcSystemPower() (watts float64, ok bool) {
	w, found := smcReadFloat("PSTR")
	if !found || w <= 0 {
		return 0, false
	}
	return w, true
}

// smcCPUPower reads the raw SMC CPU-cluster power keys — the whole CPU complex,
// not just the P-cores — which carry CPU power on Apple Silicon Pro/Max even
// where the IOReport energy model reports 0. On the M3 Max the P rails are the
// two 6-core P-clusters (PC02, PC42) and the E rails the E-core / cluster keys
// that scale with CPU load (PC03, PC43); GPU (PC1x/PC2x) and memory (PC32) keys
// are deliberately excluded. Verified live: ~20 W under a full 16-core load.
// The P/E split is kept so the caller can surface the per-cluster rails. Total
// is OK when at least one key is present (0 at idle is a real reading); the
// zero reading is returned on chips whose keys aren't mapped, so the caller
// falls back to Unavailable.
func smcCPUPower() smcCPUReading {
	var r smcCPUReading
	sum := func(keys ...string) (float64, bool) {
		var w float64
		var any bool
		for _, k := range keys {
			if v, found := smcReadFloat(k); found {
				w += v
				any = true
			}
		}
		return w, any
	}
	r.P, r.POK = sum("PC02", "PC42")
	r.E, r.EOK = sum("PC03", "PC43")
	r.Total = r.P + r.E
	r.OK = r.POK || r.EOK
	// Guard against a chip whose key layout differs: if these keys mean something
	// else there, or a float misreads, refuse an implausible CPU figure (NaN/Inf,
	// negative, or > 200 W) so we fall back to Unavailable rather than show garbage.
	if !r.OK || math.IsNaN(r.Total) || math.IsInf(r.Total, 0) || r.Total < 0 || r.Total > 200 {
		return smcCPUReading{}
	}
	return r
}
