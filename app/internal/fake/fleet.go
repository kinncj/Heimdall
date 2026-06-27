// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package fake provides a live demo fleet: a host registry that keeps a mix of
// online/stale/offline hosts fresh and lightly animated on each tick, so the
// dashboard demo doesn't age out to all-offline.
package fake

import (
	"math/rand"
	"time"

	"heimdall/app/internal/domain"
)

type spec struct {
	id, os, arch, class  string
	cpu, mem, disk, temp float64
	gpu, pwr             float64
	hasGPU, hasPower     bool
	lastSeenAgo          time.Duration // offset from now -> drives intended state
}

// Source is a live fake fleet. Tick re-observes every host (with jitter) at its
// intended age, so online stays online, one host stays stale, one stays offline.
type Source struct {
	specs []spec
	reg   *domain.HostRegistry
}

// New seeds the fleet and produces its first frame at now.
func New(now time.Time) *Source {
	s := &Source{
		reg: domain.NewHostRegistry(10*time.Second, 30*time.Second),
		specs: []spec{
			{"workstation", "linux", "amd64", "tower", 42, 68, 55, 61, 0, 0, false, false, 0},
			{"dgx-spark", "linux", "arm64", "dgx-spark", 91, 74, 40, 71, 91, 142, true, true, 0},
			{"strix-halo", "linux", "amd64", "hp", 55, 50, 33, 58, 0, 0, false, true, 0},
			{"mac-mini", "darwin", "arm64", "mac", 38, 45, 60, 49, 38, 22, true, false, 12 * time.Second},
			{"rpi-5", "linux", "arm64", "pi", 22, 30, 70, 55, 0, 0, false, false, 0},
			{"alienware", "windows", "amd64", "alien", 77, 80, 48, 68, 77, 96, true, true, 0},
			{"edge-cloud", "linux", "amd64", "cloud", 30, 41, 52, 60, 0, 0, false, false, 45 * time.Second},
		},
	}
	for _, h := range s.specs {
		s.reg.Enroll(domain.Host{
			ID: domain.HostID(h.id), Hostname: h.id, DisplayName: h.id,
			Context: domain.HostContext{OS: h.os, Arch: h.arch, Labels: map[string]string{"class": h.class}},
		}, now.Add(-time.Minute))
	}
	s.Tick(now)
	return s
}

// Registry returns the underlying host registry.
func (s *Source) Registry() *domain.HostRegistry { return s.reg }

// Tick re-observes every host at its intended age with small jitter, then
// recomputes liveness — keeping the demo live indefinitely.
func (s *Source) Tick(now time.Time) {
	for i := range s.specs {
		h := &s.specs[i]
		h.cpu = walk(h.cpu, 9)
		if h.hasGPU {
			h.gpu = walk(h.gpu, 7)
		}
		h.mem = walk(h.mem, 2)
		s.reg.Observe(domain.HostID(h.id), metricsFor(h), now.Add(-h.lastSeenAgo))
	}
	s.reg.Evaluate(now)
}

func walk(v, step float64) float64 {
	v += (rand.Float64()*2 - 1) * step
	if v < 2 {
		v = 2
	}
	if v > 99 {
		v = 99
	}
	return v
}

func metricsFor(h *spec) []domain.Metric {
	ms := []domain.Metric{
		{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: h.cpu},
		{Name: "mem.used", Unit: "percent", Status: domain.StatusOK, Gauge: h.mem},
		{Name: "disk.used", Unit: "percent", Status: domain.StatusOK, Gauge: h.disk},
		{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: h.temp},
	}
	if h.hasGPU {
		ms = append(ms, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: h.gpu})
	} else {
		ms = append(ms, domain.Metric{Name: "gpu.util", Status: domain.StatusUnavailable, Detail: "no GPU"})
	}
	if h.hasPower {
		ms = append(ms, domain.Metric{Name: "power.pkg", Unit: "watts", Status: domain.StatusOK, Gauge: h.pwr})
	} else {
		ms = append(ms, domain.Metric{Name: "power.pkg", Status: domain.StatusInsufficientPermission, Detail: "needs helper"})
	}
	return ms
}
