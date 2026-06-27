// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package alert (Gjallarhorn) evaluates threshold rules against the hub's live
// host views and raises firing/resolved alerts. It reuses the snapshots that
// already drive the liveness state machine; a breach must persist for a rule's
// For duration before it fires, so brief spikes do not flap. See ADR 0012.
package alert

import (
	"time"

	"heimdall/app/internal/domain"
)

// Op is a threshold comparison operator.
type Op string

const (
	OpGT Op = ">"
	OpLT Op = "<"
	OpGE Op = ">="
	OpLE Op = "<="
)

// Rule is one threshold over one metric, optionally scoped to hosts whose tags
// match every key in Match. A breach must hold for For before the alert fires.
type Rule struct {
	Name      string            `json:"name"`
	Metric    string            `json:"metric"`
	Op        Op                `json:"op"`
	Threshold float64           `json:"threshold"`
	For       time.Duration     `json:"for"`
	Match     map[string]string `json:"match,omitempty"`
}

// State is an alert's lifecycle state.
type State string

const (
	Firing   State = "firing"
	Resolved State = "resolved"
)

// Event is emitted on a firing/resolved transition.
type Event struct {
	Rule   string
	Host   string
	State  State
	Metric string
	Value  float64
	At     time.Time
}

// Engine tracks breach durations and active alerts across evaluations.
type Engine struct {
	rules  []Rule
	since  map[string]time.Time
	firing map[string]bool
}

// NewEngine builds an engine for the given rules.
func NewEngine(rules []Rule) *Engine {
	return &Engine{rules: rules, since: map[string]time.Time{}, firing: map[string]bool{}}
}

// Evaluate checks every rule against every matching host and returns the
// firing/resolved transitions that occurred at now.
func (e *Engine) Evaluate(views []domain.HostView, now time.Time) []Event {
	var events []Event
	for _, r := range e.rules {
		for _, v := range views {
			if !r.matches(v.Host.Context.Labels) {
				continue
			}
			key := r.Name + "\x00" + string(v.Host.ID)
			value, ok := metricValue(v, r.Metric)
			breaching := ok && breaches(value, r.Op, r.Threshold)

			if !breaching {
				delete(e.since, key)
				if e.firing[key] {
					e.firing[key] = false
					events = append(events, Event{r.Name, string(v.Host.ID), Resolved, r.Metric, value, now})
				}
				continue
			}
			if _, seen := e.since[key]; !seen {
				e.since[key] = now
			}
			if !e.firing[key] && now.Sub(e.since[key]) >= r.For {
				e.firing[key] = true
				events = append(events, Event{r.Name, string(v.Host.ID), Firing, r.Metric, value, now})
			}
		}
	}
	return events
}

// Active returns the number of currently firing alerts.
func (e *Engine) Active() int {
	n := 0
	for _, on := range e.firing {
		if on {
			n++
		}
	}
	return n
}

func (r Rule) matches(labels map[string]string) bool {
	for k, v := range r.Match {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func metricValue(v domain.HostView, name string) (float64, bool) {
	for _, m := range v.LastSnapshot {
		if m.Name == name && m.Status == domain.StatusOK && m.Kind != domain.KindPerCore {
			return m.Gauge, true
		}
	}
	return 0, false
}

func breaches(value float64, op Op, threshold float64) bool {
	switch op {
	case OpGT:
		return value > threshold
	case OpLT:
		return value < threshold
	case OpGE:
		return value >= threshold
	case OpLE:
		return value <= threshold
	}
	return false
}
