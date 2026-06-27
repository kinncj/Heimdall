// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"context"
	"fmt"
	"time"
)

// Registry holds the active adapters and collects from all of them with
// per-adapter isolation: a slow, erroring, or panicking adapter is bounded and
// surfaced as Status=Error for its own metrics only — it never drops the host
// or its sibling adapters' readings.
type Registry struct {
	adapters []Adapter
	timeout  time.Duration // per-adapter Collect deadline; <=0 disables it
}

// NewRegistry returns a Registry that bounds each adapter's Collect to timeout.
func NewRegistry(timeout time.Duration) *Registry {
	return &Registry{timeout: timeout}
}

// Register adds an adapter. Order is preserved in Collect output.
func (r *Registry) Register(a Adapter) {
	r.adapters = append(r.adapters, a)
}

// Adapters returns the number of registered adapters.
func (r *Registry) Adapters() int { return len(r.adapters) }

// Collect gathers metrics from every adapter. Each adapter runs in isolation:
// failures are converted to Error metrics and never abort the overall sweep.
func (r *Registry) Collect(ctx context.Context) []Metric {
	out := make([]Metric, 0, len(r.adapters))
	for _, a := range r.adapters {
		out = append(out, r.collectOne(ctx, a)...)
	}
	return out
}

func (r *Registry) collectOne(ctx context.Context, a Adapter) []Metric {
	info := a.Describe()

	cctx := ctx
	if r.timeout > 0 {
		var cancel context.CancelFunc
		cctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	type result struct {
		metrics []Metric
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				ch <- result{err: fmt.Errorf("adapter %q panicked: %v", info.ID, p)}
			}
		}()
		ms, err := a.Collect(cctx)
		ch <- result{metrics: ms, err: err}
	}()

	var res result
	if r.timeout > 0 {
		select {
		case res = <-ch:
		case <-cctx.Done():
			res = result{err: fmt.Errorf("adapter %q timed out after %s", info.ID, r.timeout)}
		}
	} else {
		res = <-ch
	}

	if res.err != nil {
		return errorMetrics(info, res.err)
	}
	return res.metrics
}

// errorMetrics turns an adapter failure into one Error metric per declared
// metric name, so the dashboard shows exactly which signals are faulted while
// every other adapter's metrics flow normally.
func errorMetrics(info AdapterInfo, err error) []Metric {
	now := time.Now()
	if len(info.Metrics) == 0 {
		return []Metric{{Name: info.ID, Status: StatusError, Detail: err.Error(), At: now}}
	}
	out := make([]Metric, 0, len(info.Metrics))
	for _, name := range info.Metrics {
		out = append(out, Metric{Name: name, Status: StatusError, Detail: err.Error(), At: now})
	}
	return out
}
