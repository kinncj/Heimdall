// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package alert

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func host(id string, cpu float64, labels map[string]string) domain.HostView {
	return domain.HostView{
		Host:         domain.Host{ID: domain.HostID(id), Context: domain.HostContext{Labels: labels}},
		State:        domain.StateOnline,
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: cpu}},
	}
}

func TestFiresWhenBreachingAndResolves(t *testing.T) {
	e := NewEngine([]Rule{{Name: "hot-cpu", Metric: "cpu.util", Op: OpGT, Threshold: 90}})
	t0 := time.Unix(1_000, 0)

	ev := e.Evaluate([]domain.HostView{host("a", 95, nil)}, t0)
	if len(ev) != 1 || ev[0].State != Firing {
		t.Fatalf("want one Firing event, got %+v", ev)
	}
	if e.Active() != 1 {
		t.Fatalf("active = %d, want 1", e.Active())
	}
	// re-evaluating while still breaching does not re-fire
	if ev := e.Evaluate([]domain.HostView{host("a", 95, nil)}, t0.Add(time.Second)); len(ev) != 0 {
		t.Fatalf("should not re-fire, got %+v", ev)
	}
	// dropping below threshold resolves
	ev = e.Evaluate([]domain.HostView{host("a", 50, nil)}, t0.Add(2*time.Second))
	if len(ev) != 1 || ev[0].State != Resolved || e.Active() != 0 {
		t.Fatalf("want one Resolved event and active=0, got %+v active=%d", ev, e.Active())
	}
}

func TestForDurationSuppressesSpikes(t *testing.T) {
	e := NewEngine([]Rule{{Name: "sustained", Metric: "cpu.util", Op: OpGT, Threshold: 90, For: 5 * time.Minute}})
	t0 := time.Unix(1_000, 0)

	if ev := e.Evaluate([]domain.HostView{host("a", 95, nil)}, t0); len(ev) != 0 {
		t.Fatalf("must not fire before For elapses, got %+v", ev)
	}
	// brief spike clears before For elapses -> never fires
	if ev := e.Evaluate([]domain.HostView{host("a", 10, nil)}, t0.Add(time.Minute)); len(ev) != 0 {
		t.Fatalf("spike should clear silently, got %+v", ev)
	}
	// sustained breach past For fires
	e.Evaluate([]domain.HostView{host("a", 95, nil)}, t0.Add(2*time.Minute))
	ev := e.Evaluate([]domain.HostView{host("a", 95, nil)}, t0.Add(8*time.Minute))
	if len(ev) != 1 || ev[0].State != Firing {
		t.Fatalf("sustained breach should fire, got %+v", ev)
	}
}

func TestTagScoping(t *testing.T) {
	e := NewEngine([]Rule{{Name: "prod-only", Metric: "cpu.util", Op: OpGT, Threshold: 90, Match: map[string]string{"env": "prod"}}})
	t0 := time.Unix(1_000, 0)
	views := []domain.HostView{
		host("dev", 99, map[string]string{"env": "dev"}),
		host("prod", 99, map[string]string{"env": "prod"}),
	}
	ev := e.Evaluate(views, t0)
	if len(ev) != 1 || ev[0].Host != "prod" {
		t.Fatalf("only the prod host should fire, got %+v", ev)
	}
}
