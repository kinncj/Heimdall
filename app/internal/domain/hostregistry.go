// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"sort"
	"sync"
	"time"
)

// HostRegistry is the hub-side liveness tracker. It dedupes hosts by stable
// HostID, retains the last-known snapshot through stale/offline, and drives the
// enrolling -> online -> stale -> offline -> online state machine. It is safe
// for concurrent use.
type HostRegistry struct {
	mu           sync.RWMutex
	staleAfter   time.Duration
	offlineAfter time.Duration
	purgeAfter   time.Duration
	hubLabels    map[string]string
	hosts        map[HostID]*hostEntry
}

type hostEntry struct {
	host        Host
	state       HostState
	lastSeen    time.Time
	observed    bool
	snapshot    []Metric
	labels      map[string]string
	alerts      []string
	processes   []ProcessRow
	processesAt time.Time
	logRing     []LogLine
	lastCommand *CommandResult
}

// LogRingCap bounds the per-host log ring retained for the dashboard's log modal.
const LogRingCap = 500

// NewHostRegistry returns a registry with the given liveness thresholds.
func NewHostRegistry(staleAfter, offlineAfter time.Duration) *HostRegistry {
	return &HostRegistry{
		staleAfter:   staleAfter,
		offlineAfter: offlineAfter,
		hosts:        make(map[HostID]*hostEntry),
	}
}

// SetPurgeAfter drops a host from the registry once it has been unseen for this
// long, bounding memory under host churn. Purging happens during Evaluate. Zero
// (the default) disables purging — offline hosts are retained indefinitely.
func (r *HostRegistry) SetPurgeAfter(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.purgeAfter = d
}

// Enroll registers or updates a host's identity. A returning host (same ID) is
// updated in place — never duplicated — and keeps its current liveness state
// until its next observation.
func (r *HostRegistry) Enroll(h Host, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.hosts[h.ID]; ok {
		e.host = h
		return
	}
	r.hosts[h.ID] = &hostEntry{host: h, state: StateEnrolling, lastSeen: now}
}

// SetHubLabels sets this hub's own tags (Realms). They are inherited by every
// host the hub reports, but a host's own tag of the same key wins.
func (r *HostRegistry) SetHubLabels(labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hubLabels = labels
}

// SetAlerts records the names of rules currently firing for a host (Gjallarhorn),
// so they ride into HostView and out to dashboards. An unknown host is ignored.
func (r *HostRegistry) SetAlerts(id HostID, alerts []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.hosts[id]; ok {
		e.alerts = alerts
	}
}

// RecordPush stores a host's latest pushed process table and appends pushed log
// lines to its bounded ring (Heimdallr's sight, ADR 0017). nil processes leave the
// previous table untouched; the log ring is capped at LogRingCap. It creates the
// host if unseen and never changes liveness — Observe owns that.
func (r *HostRegistry) RecordPush(id HostID, processes []ProcessRow, processesAt time.Time, logs []LogLine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.hosts[id]
	if !ok {
		e = &hostEntry{host: Host{ID: id}}
		r.hosts[id] = e
	}
	if processes != nil {
		e.processes = processes
		e.processesAt = processesAt
	}
	if len(logs) > 0 {
		e.logRing = append(e.logRing, logs...)
		if len(e.logRing) > LogRingCap {
			e.logRing = e.logRing[len(e.logRing)-LogRingCap:]
		}
	}
}

// RecordCommandResult stores a host's most recent on-demand command result (v2
// Phase 2), creating the host if unseen. It does not change liveness.
func (r *HostRegistry) RecordCommandResult(id HostID, res *CommandResult) {
	if res == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.hosts[id]
	if !ok {
		e = &hostEntry{host: Host{ID: id}}
		r.hosts[id] = e
	}
	e.lastCommand = res
}

// Observe records a fresh metric snapshot, marking the host Online and storing
// the snapshot as the last-known values. labels are the host's effective tags as
// received on the wire; nil leaves the previous labels untouched.
func (r *HostRegistry) Observe(id HostID, snapshot []Metric, labels map[string]string, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.hosts[id]
	if !ok {
		e = &hostEntry{host: Host{ID: id}}
		r.hosts[id] = e
	}
	e.observed = true
	e.lastSeen = now
	e.snapshot = snapshot
	if labels != nil {
		e.labels = labels
	}
	e.state = StateOnline
}

// Evaluate recomputes liveness for every observed host against now, and purges
// hosts unseen past the purge horizon (see SetPurgeAfter). Enrolling hosts
// (never observed) are left untouched. Last-known values are preserved until a
// host is purged.
func (r *HostRegistry) Evaluate(now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, e := range r.hosts {
		if !e.observed {
			continue
		}
		age := now.Sub(e.lastSeen)
		if r.purgeAfter > 0 && age > r.purgeAfter {
			delete(r.hosts, id)
			continue
		}
		switch {
		case age > r.offlineAfter:
			e.state = StateOffline
		case age > r.staleAfter:
			e.state = StateStale
		default:
			e.state = StateOnline
		}
	}
}

// Host returns a copy of one host's view.
func (r *HostRegistry) Host(id HostID) (HostView, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.hosts[id]
	if !ok {
		return HostView{}, false
	}
	return e.view(r.hubLabels), true
}

// Hosts returns all host views, sorted by ID for stable dashboard rendering.
func (r *HostRegistry) Hosts() []HostView {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]HostView, 0, len(r.hosts))
	for _, e := range r.hosts {
		out = append(out, e.view(r.hubLabels))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Host.ID < out[j].Host.ID })
	return out
}

// Count returns the number of registered hosts.
func (r *HostRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hosts)
}

func (e *hostEntry) view(hubLabels map[string]string) HostView {
	var snap []Metric
	if e.snapshot != nil {
		snap = append([]Metric(nil), e.snapshot...)
	}
	host := e.host
	// Precedence: hub tags < enrolled tags < live (observed) tags. The host's own
	// tags win over the hub's; the latest from the wire wins over enrollment.
	host.Context.Labels = mergeLabels(hubLabels, mergeLabels(e.host.Context.Labels, e.labels))
	var alerts []string
	if e.alerts != nil {
		alerts = append([]string(nil), e.alerts...)
	}
	var procs []ProcessRow
	if e.processes != nil {
		procs = append([]ProcessRow(nil), e.processes...)
	}
	var logs []LogLine
	if e.logRing != nil {
		logs = append([]LogLine(nil), e.logRing...)
	}
	return HostView{
		Host: host, State: e.state, LastSeen: e.lastSeen, LastSnapshot: snap, Alerts: alerts,
		Processes: procs, ProcessesAt: e.processesAt, Logs: logs, LastCommand: e.lastCommand,
	}
}

// mergeLabels combines hub-level tags with a host's own tags; the host's value
// wins on a key conflict (Realms inheritance, host overrides hub). Returns nil
// when both are empty so the common untagged case allocates nothing.
func mergeLabels(hub, host map[string]string) map[string]string {
	if len(hub) == 0 && len(host) == 0 {
		return nil
	}
	out := make(map[string]string, len(hub)+len(host))
	for k, v := range hub {
		out[k] = v
	}
	for k, v := range host {
		out[k] = v
	}
	return out
}
