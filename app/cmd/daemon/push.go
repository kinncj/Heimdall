// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/logs"
	"heimdall/app/internal/proc"
	v1 "heimdall/common/proto/monitoring/v1"
)

// pushLogCap bounds the log lines buffered between snapshots, so a noisy source
// cannot inflate a single push (Heimdallr's sight, ADR 0017 — bandwidth).
const pushLogCap = 200

// pusher accumulates the host observability the daemon pushes to the hub on its
// existing metric stream: tailed log lines and a periodic process table. The
// daemon never serves; it only pushes what it collects locally.
type pusher struct {
	mu      sync.Mutex
	pending []domain.LogLine
	dropped bool
	procs   []domain.ProcessRow
	procsAt time.Time
	// Demand window (v2, ADR 0018): the hub opens/closes pushing via StreamControl.
	// Default open so an old hub (which sends no directive) keeps v1 behaviour.
	wantLogs  bool
	wantProcs bool
	// allowCommands opts this host into on-demand command execution (v2 Phase 2).
	allowCommands bool
	// On-demand command results awaiting delivery. A FIFO queue (not a single
	// slot): the wire carries one result per snapshot, so if several commands
	// finish before the next send, each result is delivered on its own snapshot
	// rather than overwriting the previous one.
	results []*v1.ControlResponse
}

// setResult enqueues a finished on-demand command result for delivery.
func (p *pusher) setResult(r *v1.ControlResponse) {
	p.mu.Lock()
	p.results = append(p.results, r)
	p.mu.Unlock()
}

// drainResult dequeues the oldest pending command result, or nil if none.
func (p *pusher) drainResult() *v1.ControlResponse {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.results) == 0 {
		return nil
	}
	r := p.results[0]
	p.results = p.results[1:]
	return r
}

// hasResults reports whether any command results are still queued, so the send
// loop can flush them all (one snapshot each) instead of one per tick.
func (p *pusher) hasResults() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.results) > 0
}

// setWindow applies a demand directive from the hub: push logs / processes only
// while a dashboard is watching.
func (p *pusher) setWindow(logs, procs bool) {
	p.mu.Lock()
	p.wantLogs, p.wantProcs = logs, procs
	p.mu.Unlock()
}

func (p *pusher) addLine(source, line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.pending) >= pushLogCap {
		p.dropped = true
		return
	}
	ll := domain.LogLine{Source: source, At: time.Now(), Line: line}
	if p.dropped {
		ll.RateLimited = true
		p.dropped = false
	}
	p.pending = append(p.pending, ll)
}

func (p *pusher) setProcs(rows []domain.ProcessRow, at time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.procs, p.procsAt = rows, at
}

// drain returns the latest process table and the log lines accumulated since the
// last drain, which it clears — but only the parts the hub's demand window asks
// for (v2). When the window is closed, accumulated lines are dropped so they don't
// grow unbounded while nobody is watching.
func (p *pusher) drain() ([]domain.ProcessRow, time.Time, []domain.LogLine) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var procs []domain.ProcessRow
	var procsAt time.Time
	if p.wantProcs {
		procs, procsAt = p.procs, p.procsAt
	}
	var lines []domain.LogLine
	if p.wantLogs {
		lines = p.pending
	}
	p.pending = nil
	return procs, procsAt, lines
}

// startPush launches the tail and process-collection goroutines for the
// configured sources/interval. It always returns a non-nil pusher — even a daemon
// that pushes neither logs nor a process table needs one to receive directives
// (v2 Phase 2) — along with the reserved labels that advertise the capabilities to
// the hub and dashboard (`_logs`, `_proc`, `_cmd`). Reserved keys are filtered
// from user tags.
func startPush(ctx context.Context, sources logs.Sources, procInterval time.Duration, src proc.Source, allowCommands bool) (*pusher, map[string]string) {
	// Always return a pusher when streaming: even a daemon that pushes neither logs
	// nor a process table must be able to receive directives (v2 Phase 2).
	// Observability collection below is gated on its own config.
	// Open by default: a v1 hub never sends a window directive, so the daemon keeps
	// pushing per its config until a v2 hub tells it otherwise.
	p := &pusher{wantLogs: true, wantProcs: true, allowCommands: allowCommands}
	labels := map[string]string{}
	if allowCommands {
		labels["_cmd"] = "1" // advertise the on-demand command capability (gates the UI)
	}

	for alias, path := range sources {
		alias, path := alias, path
		go func() { _ = logs.Tail(ctx, path, func(line string) { p.addLine(alias, line) }) }()
	}
	if len(sources) > 0 {
		labels["_logs"] = strings.Join(sources.Aliases(), ",")
	}

	if procInterval > 0 {
		labels["_proc"] = "1"
		go func() {
			collect := func() {
				cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if rows, err := src.Collect(cctx); err == nil {
					p.setProcs(rows, time.Now())
				}
			}
			collect() // seed immediately so the first snapshot carries a table
			t := time.NewTicker(procInterval)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					collect()
				}
			}
		}()
	}
	return p, labels
}
