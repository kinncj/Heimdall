// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"testing"
	"time"
)

func sampleHost(id string) Host {
	return Host{
		ID:          HostID(id),
		Hostname:    id,
		DisplayName: id,
		Context:     HostContext{OS: "linux", Arch: "arm64", Locale: "en_US.UTF-8"},
	}
}

func snapshot() []Metric {
	return []Metric{{Name: "cpu.util", Status: StatusOK, Gauge: 33}}
}

// A freshly enrolled host is Enrolling until its first observation, then Online.
func TestEnrollThenObserveBecomesOnline(t *testing.T) {
	t0 := time.Unix(1_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)

	r.Enroll(sampleHost("dgx"), t0)
	if v, _ := r.Host("dgx"); v.State != StateEnrolling {
		t.Fatalf("want Enrolling before first observation, got %v", v.State)
	}

	r.Observe("dgx", snapshot(), t0.Add(time.Second))
	v, ok := r.Host("dgx")
	if !ok || v.State != StateOnline {
		t.Fatalf("want Online after observation, got %v (ok=%v)", v.State, ok)
	}
	if len(v.LastSnapshot) != 1 || v.LastSnapshot[0].Gauge != 33 {
		t.Errorf("last-known snapshot not retained: %+v", v.LastSnapshot)
	}
}

// Past stale_after with no updates: Stale, but last-known values + timestamp stay.
func TestGoesStaleButRetainsLastKnown(t *testing.T) {
	t0 := time.Unix(2_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	r.Enroll(sampleHost("rpi"), t0)
	r.Observe("rpi", snapshot(), t0)

	r.Evaluate(t0.Add(15 * time.Second)) // > stale_after, < offline_after
	v, _ := r.Host("rpi")
	if v.State != StateStale {
		t.Fatalf("want Stale, got %v", v.State)
	}
	if len(v.LastSnapshot) != 1 {
		t.Errorf("stale host must keep last-known values, got %+v", v.LastSnapshot)
	}
	if !v.LastSeen.Equal(t0) {
		t.Errorf("LastSeen must reflect the last real update %v, got %v", t0, v.LastSeen)
	}
}

// Past offline_after: Offline.
func TestGoesOfflineAfterThreshold(t *testing.T) {
	t0 := time.Unix(3_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	r.Enroll(sampleHost("mac"), t0)
	r.Observe("mac", snapshot(), t0)

	r.Evaluate(t0.Add(45 * time.Second))
	if v, _ := r.Host("mac"); v.State != StateOffline {
		t.Fatalf("want Offline, got %v", v.State)
	}
}

// Reconnect after an outage returns to Online under the SAME id with no
// duplicate host created (story 0001).
func TestReconnectResumesSameHostNoDuplicate(t *testing.T) {
	t0 := time.Unix(4_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	r.Enroll(sampleHost("alien"), t0)
	r.Observe("alien", snapshot(), t0)
	r.Evaluate(t0.Add(45 * time.Second)) // offline

	// network returns; daemon reconnects and re-enrolls with the same stable id
	r.Enroll(sampleHost("alien"), t0.Add(60*time.Second))
	r.Observe("alien", snapshot(), t0.Add(61*time.Second))

	if n := r.Count(); n != 1 {
		t.Fatalf("reconnect must not duplicate host registration, count=%d", n)
	}
	if v, _ := r.Host("alien"); v.State != StateOnline {
		t.Errorf("want Online after reconnect, got %v", v.State)
	}
}

// A host unseen past the purge horizon is removed entirely, so the registry —
// and the trend buffers keyed off it — never grow without bound under churn
// (architecture.md risk: in-memory ring buffers grow with fleet size).
func TestPurgesLongOfflineHost(t *testing.T) {
	t0 := time.Unix(5_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	r.SetPurgeAfter(5 * time.Minute)
	r.Enroll(sampleHost("ephemeral"), t0)
	r.Observe("ephemeral", snapshot(), t0)

	r.Evaluate(t0.Add(45 * time.Second)) // offline, still retained for visibility
	if r.Count() != 1 {
		t.Fatalf("offline host should be retained, count=%d", r.Count())
	}
	r.Evaluate(t0.Add(6 * time.Minute)) // past the purge horizon
	if r.Count() != 0 {
		t.Fatalf("long-offline host should be purged, count=%d", r.Count())
	}
}

// Enrolling the same id twice updates identity in place — never two entries.
func TestEnrollSameIDDeduplicates(t *testing.T) {
	t0 := time.Unix(5_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	r.Enroll(sampleHost("box"), t0)

	updated := sampleHost("box")
	updated.DisplayName = "workstation"
	r.Enroll(updated, t0.Add(time.Second))

	if n := r.Count(); n != 1 {
		t.Fatalf("want 1 host after re-enroll, got %d", n)
	}
	if v, _ := r.Host("box"); v.Host.DisplayName != "workstation" {
		t.Errorf("re-enroll should update identity, got %q", v.Host.DisplayName)
	}
}

// Hosts() returns a stable, id-sorted view for the dashboard.
func TestHostsSortedByID(t *testing.T) {
	t0 := time.Unix(6_000, 0)
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	for _, id := range []string{"ceta", "alfa", "beta"} {
		r.Enroll(sampleHost(id), t0)
	}
	got := r.Hosts()
	if len(got) != 3 || got[0].Host.ID != "alfa" || got[2].Host.ID != "ceta" {
		t.Fatalf("want id-sorted hosts [alfa beta ceta], got %v", []HostID{got[0].Host.ID, got[1].Host.ID, got[2].Host.ID})
	}
}
