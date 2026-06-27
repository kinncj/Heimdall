// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import "time"

// HostID is the stable, content-derived identity of a monitored machine. It is
// stable across restarts and reconnects so a returning daemon re-registers as
// the same host rather than creating a duplicate.
type HostID string

// HostContext holds slow-changing descriptive attributes of a host. Sent on
// enrollment and refreshed on change, not on every metric sample.
type HostContext struct {
	OS           string
	OSVersion    string
	Arch         string
	Kernel       string
	Locale       string
	Timezone     string
	BootTimeUnix int64
	AgentVersion string
	Labels       map[string]string
}

// Host is a monitored machine's identity.
type Host struct {
	ID          HostID
	Hostname    string
	DisplayName string
	Context     HostContext
}

// HostState is the liveness state of a host as seen by the hub.
type HostState int

const (
	StateUnknown   HostState = iota
	StateEnrolling           // registered, not yet streaming
	StateOnline              // streaming within the freshness window
	StateStale               // no updates past stale_after; last-known retained
	StateOffline             // no updates past offline_after
)

func (s HostState) String() string {
	switch s {
	case StateEnrolling:
		return "enrolling"
	case StateOnline:
		return "online"
	case StateStale:
		return "stale"
	case StateOffline:
		return "offline"
	default:
		return "unknown"
	}
}

// HostView is an immutable snapshot of a host's liveness for the dashboard.
// LastSnapshot holds the last-known metric set, retained through stale/offline
// so the UI never blanks out a machine that has merely gone quiet.
type HostView struct {
	Host         Host
	State        HostState
	LastSeen     time.Time
	LastSnapshot []Metric
	Alerts       []string // names of rules currently firing for this host (Gjallarhorn)
}
