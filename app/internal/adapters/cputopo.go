// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"strconv"
	"strings"
	"sync"

	"heimdall/app/internal/domain"
)

// Core-type ids are the standardized, platform-neutral taxonomy — defined once
// in the domain (domain.CorePerf/CoreEff/CoreLP). The probes below translate
// each platform's own scheme into these ids; nothing here re-defines them.
const (
	corePerf = domain.CorePerf
	coreEff  = domain.CoreEff
	coreLP   = domain.CoreLP
)

// topoOnce caches the probe: core topology cannot change while the daemon runs,
// and the sysctl/sysfs/syscall reads don't need repeating every collect.
var (
	topoOnce  sync.Once
	topoTypes []int
	topoOK    bool
)

// coreTopologyMetric returns the cpu.topology metric for a host with `total`
// logical cores, or ok=false when no probe applies (the adapter then omits the
// metric entirely — the renderer falls back to the unlabelled grid).
func coreTopologyMetric(total int) (domain.Metric, bool) {
	topoOnce.Do(func() {
		topoTypes, topoOK = coreTypes(total)
	})
	if !topoOK || len(topoTypes) != total {
		return domain.Metric{}, false
	}
	return topologyMetric(topoTypes), true
}

// topologyMetric encodes a core-type slice as the cpu.topology metric:
// PerCore[i] is the type id of logical core i, Detail is the standardized human
// summary. Gauge is deliberately left 0 — per-core metrics ride the proto
// per_core oneof, so a Gauge would be dropped on the wire anyway; consumers read
// the type layout from PerCore via domain.CoreGroups.
func topologyMetric(types []int) domain.Metric {
	per := make([]float64, len(types))
	for i, t := range types {
		per[i] = float64(t)
	}
	return domain.Metric{
		Name: "cpu.topology", Unit: "type", Status: domain.StatusOK,
		Kind: domain.KindPerCore, PerCore: per,
		Detail: domain.CoreTypeSummary(per),
	}
}

// typesFromCounts builds a type slice from per-tier core counts (the macOS
// sysctl shape). eFirst says whether the efficiency cluster owns the low
// logical ids (true on Apple Silicon — verified live on an M3 Max: background-
// QoS load lands on c0–c3). Counts that don't add up are refused.
func typesFromCounts(p, e, total int, eFirst bool) ([]int, bool) {
	if total <= 0 || p < 0 || e < 0 || p+e != total {
		return nil, false
	}
	out := make([]int, 0, total)
	first, second, firstType, secondType := p, e, corePerf, coreEff
	if eFirst {
		first, second, firstType, secondType = e, p, coreEff, corePerf
	}
	for i := 0; i < first; i++ {
		out = append(out, firstType)
	}
	for i := 0; i < second; i++ {
		out = append(out, secondType)
	}
	return out, true
}

// typesFromRanges builds a type slice from index-range lists (the Linux
// /sys/devices/cpu_core/cpus + cpu_atom/cpus shape, e.g. "0-15" and "16-23",
// possibly comma-separated). Every core must be claimed by exactly one list and
// stay in bounds, else the probe is refused.
func typesFromRanges(perf, eff string, total int) ([]int, bool) {
	if total <= 0 {
		return nil, false
	}
	out := make([]int, total)
	for i := range out {
		out[i] = -1
	}
	assign := func(spec string, t int) bool {
		idx, ok := parseIndexRanges(spec)
		if !ok || len(idx) == 0 {
			return false
		}
		for _, i := range idx {
			if i < 0 || i >= total || out[i] != -1 {
				return false
			}
			out[i] = t
		}
		return true
	}
	if !assign(perf, corePerf) || !assign(eff, coreEff) {
		return nil, false
	}
	for _, t := range out {
		if t == -1 {
			return nil, false
		}
	}
	return out, true
}

// parseIndexRanges parses a sysfs cpu list ("0-15", "0-3,8-11", "7").
func parseIndexRanges(spec string) ([]int, bool) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, false
	}
	var out []int
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		lo, hi := part, part
		if a, b, found := strings.Cut(part, "-"); found {
			lo, hi = a, b
		}
		l, err1 := strconv.Atoi(lo)
		h, err2 := strconv.Atoi(hi)
		if err1 != nil || err2 != nil || h < l {
			return nil, false
		}
		for i := l; i <= h; i++ {
			out = append(out, i)
		}
	}
	return out, true
}

// typesFromEfficiencyClasses maps Windows per-core EfficiencyClass values
// (ordered by logical core index; a core's class repeats for each of its
// logical CPUs) to type ids. Windows counts efficiency class upward with
// performance — the highest class is the P tier — while our ids count downward
// from P=0, so the mapping inverts.
func typesFromEfficiencyClasses(classes []int) ([]int, bool) {
	if len(classes) == 0 {
		return nil, false
	}
	max := classes[0]
	for _, c := range classes {
		if c < 0 {
			return nil, false
		}
		if c > max {
			max = c
		}
	}
	out := make([]int, len(classes))
	for i, c := range classes {
		t := max - c
		if domain.CoreTypeLabel(t) == "" {
			// more tiers than the standard taxonomy labels — refuse rather than mislabel
			return nil, false
		}
		out[i] = t
	}
	return out, true
}

// uniformTypes is the no-probe fallback: every core the same (performance) type.
func uniformTypes(total int) []int {
	out := make([]int, total)
	return out
}

// relationProcessorCore is the RelationProcessorCore record tag in Windows'
// GetLogicalProcessorInformationEx output. The parser lives here (pure bytes,
// no syscalls) so it tests on every platform; only the syscall is build-tagged.
const relationProcessorCore = 0

// efficiencyClassesFromBuffer walks SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX
// records (RelationProcessorCore) and maps each logical CPU (group-0 mask bit)
// to its core's EfficiencyClass. Record layout: Relationship u32, Size u32,
// then PROCESSOR_RELATIONSHIP{Flags u8, EfficiencyClass u8, Reserved [20]u8,
// GroupCount u16, GroupMask []GROUP_AFFINITY{Mask u64, Group u16, ...}}.
// Multi-group machines (>64 logical CPUs) are refused — they don't ship hybrid
// silicon, and an unlabelled grid beats a mislabelled one.
func efficiencyClassesFromBuffer(buf []byte, total int) ([]int, bool) {
	if total <= 0 || total > 64 {
		return nil, false
	}
	classes := make([]int, total)
	for i := range classes {
		classes[i] = -1
	}
	le := func(b []byte) uint64 {
		var v uint64
		for i := len(b) - 1; i >= 0; i-- {
			v = v<<8 | uint64(b[i])
		}
		return v
	}
	for off := 0; off+32 <= len(buf); {
		rel := uint32(le(buf[off : off+4]))
		size := int(le(buf[off+4 : off+8]))
		if size <= 0 || off+size > len(buf) {
			return nil, false
		}
		if rel == relationProcessorCore {
			effClass := int(buf[off+9])
			groupCount := int(le(buf[off+30 : off+32]))
			maskOff := off + 32
			if groupCount != 1 || maskOff+16 > off+size {
				return nil, false
			}
			mask := le(buf[maskOff : maskOff+8])
			group := int(le(buf[maskOff+8 : maskOff+10]))
			if group != 0 {
				return nil, false
			}
			for bit := 0; bit < 64; bit++ {
				if mask&(1<<bit) == 0 {
					continue
				}
				if bit >= total || classes[bit] != -1 {
					return nil, false
				}
				classes[bit] = effClass
			}
		}
		off += size
	}
	for _, c := range classes {
		if c == -1 {
			return nil, false
		}
	}
	return classes, true
}
