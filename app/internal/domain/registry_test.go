// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeAdapter is a controllable Adapter for tests.
type fakeAdapter struct {
	info     AdapterInfo
	metrics  []Metric
	err      error
	panicMsg string
	delay    time.Duration
}

func (f fakeAdapter) Describe() AdapterInfo { return f.info }

func (f fakeAdapter) Collect(ctx context.Context) ([]Metric, error) {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.panicMsg != "" {
		panic(f.panicMsg)
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.metrics, nil
}

func okAdapter(id string, metricNames []string, ms ...Metric) fakeAdapter {
	return fakeAdapter{info: AdapterInfo{ID: id, Metrics: metricNames}, metrics: ms}
}

func metricsByName(ms []Metric) map[string]Metric {
	out := make(map[string]Metric, len(ms))
	for _, m := range ms {
		out[m.Name] = m
	}
	return out
}

func TestRegistryCollectsFromAllAdapters(t *testing.T) {
	r := NewRegistry(time.Second)
	r.Register(okAdapter("cpu", []string{"cpu.util"}, Metric{Name: "cpu.util", Status: StatusOK, Gauge: 42}))
	r.Register(okAdapter("mem", []string{"mem.used"}, Metric{Name: "mem.used", Status: StatusOK, Gauge: 70}))

	got := metricsByName(r.Collect(context.Background()))
	if len(got) != 2 {
		t.Fatalf("want 2 metrics, got %d: %v", len(got), got)
	}
	if got["cpu.util"].Status != StatusOK || got["cpu.util"].Gauge != 42 {
		t.Errorf("cpu.util wrong: %+v", got["cpu.util"])
	}
	if got["mem.used"].Gauge != 70 {
		t.Errorf("mem.used wrong: %+v", got["mem.used"])
	}
}

// One adapter erroring must not drop the others, and the failing adapter's
// declared metrics surface as Status=Error (failure isolation, story 0003).
func TestAdapterErrorIsIsolated(t *testing.T) {
	r := NewRegistry(time.Second)
	r.Register(okAdapter("cpu", []string{"cpu.util"}, Metric{Name: "cpu.util", Status: StatusOK, Gauge: 10}))
	r.Register(fakeAdapter{info: AdapterInfo{ID: "temp", Metrics: []string{"temp.pkg"}}, err: errors.New("sensor offline")})

	got := metricsByName(r.Collect(context.Background()))
	if got["cpu.util"].Status != StatusOK {
		t.Errorf("healthy adapter dropped or altered: %+v", got["cpu.util"])
	}
	tm, ok := got["temp.pkg"]
	if !ok {
		t.Fatalf("failed adapter's declared metric missing; want an Error metric")
	}
	if tm.Status != StatusError {
		t.Errorf("want StatusError for failed adapter metric, got %v", tm.Status)
	}
	if tm.Detail == "" {
		t.Errorf("want a Detail explaining the failure")
	}
}

// A panicking adapter must be recovered and isolated, never crashing Collect.
func TestAdapterPanicIsIsolated(t *testing.T) {
	r := NewRegistry(time.Second)
	r.Register(okAdapter("cpu", []string{"cpu.util"}, Metric{Name: "cpu.util", Status: StatusOK, Gauge: 5}))
	r.Register(fakeAdapter{info: AdapterInfo{ID: "gpu", Metrics: []string{"gpu.util"}}, panicMsg: "nvml boom"})

	var got map[string]Metric
	done := make(chan struct{})
	go func() {
		got = metricsByName(r.Collect(context.Background()))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Collect hung or crashed on a panicking adapter")
	}
	if got["cpu.util"].Status != StatusOK {
		t.Errorf("healthy adapter dropped after sibling panic: %+v", got["cpu.util"])
	}
	if got["gpu.util"].Status != StatusError {
		t.Errorf("panicking adapter metric should be StatusError, got %v", got["gpu.util"].Status)
	}
}

// Non-OK, non-error statuses (Unavailable, InsufficientPermission) are first-class
// and must be preserved verbatim — they are NOT failures (stories 0003, 0005).
func TestNonOKStatusPreserved(t *testing.T) {
	r := NewRegistry(time.Second)
	r.Register(okAdapter("gpu", []string{"gpu.util"},
		Metric{Name: "gpu.util", Status: StatusUnavailable, Detail: "no NVIDIA GPU"}))
	r.Register(okAdapter("power", []string{"power.pkg"},
		Metric{Name: "power.pkg", Status: StatusInsufficientPermission, Detail: "needs helper"}))

	got := metricsByName(r.Collect(context.Background()))
	if got["gpu.util"].Status != StatusUnavailable {
		t.Errorf("Unavailable coerced: %+v", got["gpu.util"])
	}
	if got["power.pkg"].Status != StatusInsufficientPermission {
		t.Errorf("InsufficientPermission coerced: %+v", got["power.pkg"])
	}
}

// A slow adapter exceeding the per-adapter timeout is isolated as Error.
func TestAdapterTimeoutIsIsolated(t *testing.T) {
	r := NewRegistry(50 * time.Millisecond)
	r.Register(okAdapter("cpu", []string{"cpu.util"}, Metric{Name: "cpu.util", Status: StatusOK, Gauge: 1}))
	r.Register(fakeAdapter{info: AdapterInfo{ID: "slow", Metrics: []string{"slow.metric"}}, delay: time.Second})

	start := time.Now()
	got := metricsByName(r.Collect(context.Background()))
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("Collect did not honor per-adapter timeout, took %v", elapsed)
	}
	if got["cpu.util"].Status != StatusOK {
		t.Errorf("healthy adapter dropped on sibling timeout")
	}
	if got["slow.metric"].Status != StatusError {
		t.Errorf("timed-out adapter should be StatusError, got %v", got["slow.metric"].Status)
	}
}
