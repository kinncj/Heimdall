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
// configured sources/interval. It returns the pusher (nil when nothing is
// configured) and the reserved labels that advertise the capability to the hub
// and dashboard (`_logs`, `_proc`). Reserved keys are filtered from user tags.
func startPush(ctx context.Context, sources logs.Sources, procInterval time.Duration, src proc.Source) (*pusher, map[string]string) {
	if len(sources) == 0 && procInterval <= 0 {
		return nil, nil
	}
	// Open by default: a v1 hub never sends a window directive, so the daemon keeps
	// pushing per its config until a v2 hub tells it otherwise.
	p := &pusher{wantLogs: true, wantProcs: true}
	labels := map[string]string{}

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
