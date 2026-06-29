// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package fake provides a live demo fleet: a host registry that keeps a mix of
// online/stale/offline hosts fresh and lightly animated on each tick, so the
// dashboard demo doesn't age out to all-offline.
package fake

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"heimdall/app/internal/domain"
)

// Demo observability (Heimdallr's sight): every demo host advertises a process
// table and log sources, so every feature — logs (l, with / search), top (t, with
// s sort), grouping, filtering, alerts — is explorable on any host in --demo.
var demoLogs = map[string]string{
	"workstation": "app,sys",
	"dgx-spark":   "train,app",
	"strix-halo":  "app,sys",
	"mac-mini":    "app",
	"rpi-5":       "sensor",
	"alienware":   "game,app",
	"edge-cloud":  "app,nginx",
}

// demoProc reports whether a demo host pushes a process table — all of them do.
func demoProc(id string) bool { return true }

type spec struct {
	id, os, arch, class  string
	cpu, mem, disk, temp float64
	gpu, pwr             float64
	hasGPU, hasPower     bool
	lastSeenAgo          time.Duration // offset from now -> drives intended state
	hub, env, role       string        // Realms/Yggdrasil: origin hub + tags
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
			{"workstation", "linux", "amd64", "tower", 42, 68, 55, 61, 0, 0, false, false, 0, "home", "prod", "server"},
			{"dgx-spark", "linux", "arm64", "dgx-spark", 91, 74, 40, 71, 91, 142, true, true, 0, "home", "prod", "gpu"},
			{"strix-halo", "linux", "amd64", "hp", 55, 50, 33, 58, 0, 0, false, true, 0, "home", "dev", "server"},
			{"mac-mini", "darwin", "arm64", "mac", 38, 45, 60, 49, 38, 22, true, false, 12 * time.Second, "remote-work-station", "dev", "workstation"},
			{"rpi-5", "linux", "arm64", "pi", 22, 30, 70, 55, 0, 0, false, false, 0, "home", "dev", "edge"},
			{"alienware", "windows", "amd64", "alien", 77, 80, 48, 68, 77, 96, true, true, 0, "remote-work-station", "prod", "workstation"},
			{"edge-cloud", "linux", "amd64", "cloud", 30, 41, 52, 60, 0, 0, false, false, 45 * time.Second, "central", "prod", "edge"},
		},
	}
	for _, h := range s.specs {
		labels := map[string]string{"class": h.class, "hub": h.hub, "env": h.env, "role": h.role}
		if v := demoLogs[h.id]; v != "" {
			labels["_logs"] = v
		}
		if demoProc(h.id) {
			labels["_proc"] = "1"
		}
		s.reg.Enroll(domain.Host{
			ID: domain.HostID(h.id), Hostname: h.id, DisplayName: h.id,
			Context: domain.HostContext{OS: h.os, Arch: h.arch, Labels: labels},
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
		s.reg.Observe(domain.HostID(h.id), metricsFor(h), nil, now.Add(-h.lastSeenAgo))
		// Simulate Gjallarhorn: a host running hot raises an alert, so the badge
		// and alert count are visible in --demo without a hub or rules file.
		var alerts []string
		if h.cpu > 85 {
			alerts = []string{"cpu.util>85"}
		}
		s.reg.SetAlerts(domain.HostID(h.id), alerts)
		s.pushObservability(h, now)
	}
	s.reg.Evaluate(now)
}

// pushObservability feeds a synthetic process table and a fresh log line into the
// registry for hosts that advertise the capability, so the l/top modals animate.
func (s *Source) pushObservability(h *spec, now time.Time) {
	var procs []domain.ProcessRow
	if demoProc(h.id) {
		procs = demoProcs(h)
	}
	var lines []domain.LogLine
	for _, alias := range strings.Split(demoLogs[h.id], ",") {
		if alias == "" {
			continue
		}
		lines = append(lines, domain.LogLine{Source: alias, At: now, Line: demoLogLine(h, now)})
	}
	if procs != nil || lines != nil {
		s.reg.RecordPush(domain.HostID(h.id), procs, now, lines)
	}
}

func demoProcs(h *spec) []domain.ProcessRow {
	names := []string{"heimdall-daemon", "systemd", "sshd", h.class + "-worker", "node", "postgres", "containerd"}
	rows := make([]domain.ProcessRow, len(names))
	for i, n := range names {
		rows[i] = domain.ProcessRow{
			PID:     uint32(100 + i*7),
			PPID:    1,
			CPUPct:  walk(h.cpu/float64(i+2), 3),
			MemPct:  float64((i*3+5)%40) + 0.5,
			Command: n,
		}
	}
	return rows
}

func demoLogLine(h *spec, now time.Time) string {
	msgs := []string{"request handled", "cache warmed", "watch over all realms", "gc pause 2ms", "peer connected"}
	return fmt.Sprintf("level=info host=%s cpu=%.0f%% %s", h.id, h.cpu, msgs[int(now.Unix())%len(msgs)])
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
