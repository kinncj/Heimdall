// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package observe

import (
	"sync"
	"time"

	"heimdall/app/internal/domain"
)

// Sample is one historical reading.
type Sample struct {
	TS    time.Time
	Value float64
}

// History keeps a bounded ring of recent OK gauge samples per host+metric. It is
// in-memory only — history is lost on hub restart, by design (ADR 0011). Safe for
// concurrent use.
type History struct {
	mu     sync.Mutex
	depth  int
	series map[string][]Sample
}

// NewHistory returns a history that keeps at most depth samples per series.
func NewHistory(depth int) *History {
	if depth < 1 {
		depth = 1
	}
	return &History{depth: depth, series: make(map[string][]Sample)}
}

// Record appends the OK single-gauge metrics from each view at time now. Per-core
// and non-OK metrics are skipped — only scalar trends are retained.
func (h *History) Record(views []domain.HostView, now time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, v := range views {
		for _, m := range v.LastSnapshot {
			if m.Status != domain.StatusOK || m.Kind == domain.KindPerCore || m.Unit == "info" {
				continue
			}
			key := string(v.Host.ID) + "\x00" + m.Name
			s := append(h.series[key], Sample{TS: now, Value: m.Gauge})
			if len(s) > h.depth {
				s = s[len(s)-h.depth:]
			}
			h.series[key] = s
		}
	}
}

// Series returns a copy of the retained samples for one host+metric, oldest first.
func (h *History) Series(host, metric string) []Sample {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.series[host+"\x00"+metric]
	if s == nil {
		return nil
	}
	return append([]Sample(nil), s...)
}
