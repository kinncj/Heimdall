// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build darwin && cgo

package helper

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit -lIOReport
#include <CoreFoundation/CoreFoundation.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>

// IOReport is a private Apple framework. These energy-model counters are
// readable without root, which is how we obtain GPU (and, where exposed, CPU/ANE)
// power without elevation.
extern CFDictionaryRef IOReportCopyChannelsInGroup(CFStringRef, CFStringRef, uint64_t, uint64_t, uint64_t);
extern void IOReportMergeChannels(CFDictionaryRef, CFDictionaryRef, CFTypeRef);
extern void *IOReportCreateSubscription(void *, CFMutableDictionaryRef, CFMutableDictionaryRef *, uint64_t, CFTypeRef);
extern CFDictionaryRef IOReportCreateSamples(void *, CFMutableDictionaryRef, CFTypeRef);
extern CFDictionaryRef IOReportCreateSamplesDelta(CFDictionaryRef, CFDictionaryRef, CFTypeRef);
extern CFStringRef IOReportChannelGetChannelName(CFDictionaryRef);
extern CFStringRef IOReportChannelGetUnitLabel(CFDictionaryRef);
extern int64_t IOReportSimpleGetIntegerValue(CFDictionaryRef, int32_t);
extern int32_t IOReportStateGetCount(CFDictionaryRef);
extern CFStringRef IOReportStateGetNameForIndex(CFDictionaryRef, int32_t);
extern int64_t IOReportStateGetResidency(CFDictionaryRef, int32_t);

static void *g_sub = NULL;
static CFMutableDictionaryRef g_chan = NULL;

static int hm_init(void) {
	if (g_sub) return 0;
	CFDictionaryRef e = IOReportCopyChannelsInGroup(CFSTR("Energy Model"), NULL, 0, 0, 0);
	if (!e) return -1;
	CFDictionaryRef g = IOReportCopyChannelsInGroup(CFSTR("GPU Stats"), NULL, 0, 0, 0);
	if (g) { IOReportMergeChannels(e, g, NULL); CFRelease(g); }
	g_chan = CFDictionaryCreateMutableCopy(kCFAllocatorDefault, 0, e);
	CFRelease(e);
	if (!g_chan) return -2;
	CFMutableDictionaryRef subbed = NULL;
	g_sub = IOReportCreateSubscription(NULL, g_chan, &subbed, 0, NULL);
	if (!g_sub) { CFRelease(g_chan); g_chan = NULL; return -3; }
	return 0;
}

// hm_energy_to_watts converts an energy delta in the channel's unit, accumulated
// over sec seconds, into watts.
static double hm_energy_to_watts(int64_t v, CFStringRef u, double sec) {
	if (sec <= 0) return 0;
	char unit[32] = {0};
	if (u) CFStringGetCString(u, unit, sizeof(unit), kCFStringEncodingUTF8);
	for (int i = 0; unit[i]; i++) if (unit[i] == ' ') unit[i] = '\0';
	double joules;
	if (!strcmp(unit, "mJ")) joules = (double)v / 1e3;
	else if (!strcmp(unit, "uJ")) joules = (double)v / 1e6;
	else if (!strcmp(unit, "nJ")) joules = (double)v / 1e9;
	else joules = (double)v / 1e6;
	return joules / sec;
}

// hm_sample_power samples the energy-model counters over durationMs and writes
// CPU/GPU/ANE power in watts plus GPU utilisation (percent active residency from
// the GPU Performance States channel; -1 when unavailable). Channels read 0 on
// SoCs that do not expose them (notably CPU package power on several M-series
// chips). Returns 0 on success.
static int hm_sample_power(int durationMs, double *cpuOut, double *gpuOut, double *aneOut, double *gpuUtilOut) {
	*cpuOut = 0; *gpuOut = 0; *aneOut = 0; *gpuUtilOut = -1;
	if (hm_init() != 0) return -1;
	@autoreleasepool {
		CFDictionaryRef s1 = IOReportCreateSamples(g_sub, g_chan, NULL);
		if (!s1) return -2;
		usleep((useconds_t)durationMs * 1000);
		CFDictionaryRef s2 = IOReportCreateSamples(g_sub, g_chan, NULL);
		if (!s2) { CFRelease(s1); return -3; }
		double sec = (double)durationMs / 1000.0;
		CFDictionaryRef d = IOReportCreateSamplesDelta(s1, s2, NULL);
		CFRelease(s1); CFRelease(s2);
		if (!d) return -4;
		CFArrayRef arr = CFDictionaryGetValue(d, CFSTR("IOReportChannels"));
		CFIndex n = arr ? CFArrayGetCount(arr) : 0;
		double cpuTotal = 0, cpuTyped = 0, gpuNamed = 0, gpuAlias = 0, ane = 0;
		for (CFIndex i = 0; i < n; i++) {
			CFDictionaryRef ch = (CFDictionaryRef)CFArrayGetValueAtIndex(arr, i);
			CFStringRef nr = IOReportChannelGetChannelName(ch);
			if (!nr) continue;
			char nm[256] = {0};
			CFStringGetCString(nr, nm, sizeof(nm), kCFStringEncodingUTF8);

			// GPU utilisation: active residency over the sample window, from the
			// GPU Performance States channel (state "OFF" is idle).
			if (strcmp(nm, "GPUPH") == 0) {
				int32_t sc = IOReportStateGetCount(ch);
				int64_t total = 0, off = 0;
				for (int32_t s = 0; s < sc; s++) {
					int64_t r = IOReportStateGetResidency(ch, s);
					if (r < 0) r = 0;
					total += r;
					char st[32] = {0};
					CFStringRef sn = IOReportStateGetNameForIndex(ch, s);
					if (sn) CFStringGetCString(sn, st, sizeof(st), kCFStringEncodingUTF8);
					if (!strcmp(st, "OFF")) off += r;
				}
				if (total > 0) *gpuUtilOut = (double)(total - off) / (double)total * 100.0;
				continue;
			}

			CFStringRef ur = IOReportChannelGetUnitLabel(ch);
			int64_t v = IOReportSimpleGetIntegerValue(ch, 0);
			double w = hm_energy_to_watts(v, ur, sec);
			int typed = strstr(nm, "ECPU Energy") || strstr(nm, "PCPU Energy") || strstr(nm, "MCPU Energy") ||
			            strstr(nm, "eCPUs Energy") || strstr(nm, "pCPUs Energy") || strstr(nm, "mCPUs Energy");
			if (typed) cpuTyped += w;
			else if (strstr(nm, "CPU Energy")) cpuTotal += w;
			else if (!strcmp(nm, "GPU Energy")) gpuNamed += w;
			else if (!strcmp(nm, "GPU")) gpuAlias += w;
			else if (strstr(nm, "ANE") && strstr(nm, "Energy")) ane += w;
		}
		CFRelease(d);
		*cpuOut = (cpuTotal > 0) ? cpuTotal : cpuTyped;
		*gpuOut = (gpuNamed > 0) ? gpuNamed : gpuAlias;
		*aneOut = ane;
	}
	return 0;
}
*/
import "C"

import "sync"

var (
	ioReportMu   sync.Mutex
	ioReportPrim sync.Once
)

// ioReportPower samples Apple Silicon energy counters via the IOReport private
// framework (no root required) and returns CPU/GPU/ANE power in watts plus GPU
// utilisation as a percentage (-1 when unavailable). A power channel reads 0
// where the SoC does not expose it. ok is false only if IOReport itself is
// unavailable (e.g. an Intel Mac without the Energy Model group).
func ioReportPower(durationMs int) (cpu, gpu, ane, gpuUtil float64, ok bool) {
	ioReportMu.Lock()
	defer ioReportMu.Unlock()

	// The first delta after subscription is a warm-up artifact; prime once.
	ioReportPrim.Do(func() {
		var c, g, a, u C.double
		C.hm_sample_power(40, &c, &g, &a, &u)
	})

	var c, g, a, u C.double
	if C.hm_sample_power(C.int(durationMs), &c, &g, &a, &u) != 0 {
		return 0, 0, 0, -1, false
	}
	return float64(c), float64(g), float64(a), float64(u), true
}
