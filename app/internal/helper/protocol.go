// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package helper is the optional privileged metrics unit. heimdall-helper runs
// as a separate privileged process and serves power, GPU, and full thermal
// readings to the unprivileged daemon over a local unix socket. The daemon
// never invokes sudo or runs as root: it only reads from the socket.
package helper

import (
	"bufio"
	"encoding/json"
	"io"

	"heimdall/app/internal/domain"
)

const protocolVersion = 1

type wireMetric struct {
	Name   string  `json:"name"`
	Unit   string  `json:"unit,omitempty"`
	Status string  `json:"status"`
	Gauge  float64 `json:"gauge,omitempty"`
	Detail string  `json:"detail,omitempty"`
}

type envelope struct {
	V       int          `json:"v"`
	Metrics []wireMetric `json:"metrics"`
}

func encodeMetrics(w io.Writer, ms []domain.Metric) error {
	env := envelope{V: protocolVersion, Metrics: make([]wireMetric, 0, len(ms))}
	for _, m := range ms {
		env.Metrics = append(env.Metrics, wireMetric{
			Name:   m.Name,
			Unit:   m.Unit,
			Status: m.Status.String(),
			Gauge:  m.Gauge,
			Detail: m.Detail,
		})
	}
	return json.NewEncoder(w).Encode(env)
}

func decodeMetrics(r io.Reader) ([]domain.Metric, error) {
	var env envelope
	if err := json.NewDecoder(bufio.NewReader(r)).Decode(&env); err != nil {
		return nil, err
	}
	out := make([]domain.Metric, 0, len(env.Metrics))
	for _, wm := range env.Metrics {
		out = append(out, domain.Metric{
			Name:   wm.Name,
			Unit:   wm.Unit,
			Status: statusFromString(wm.Status),
			Gauge:  wm.Gauge,
			Detail: wm.Detail,
		})
	}
	return out, nil
}

func statusFromString(s string) domain.MetricStatus {
	switch s {
	case "ok":
		return domain.StatusOK
	case "unavailable":
		return domain.StatusUnavailable
	case "insufficient_permission":
		return domain.StatusInsufficientPermission
	case "error":
		return domain.StatusError
	default:
		return domain.StatusUnspecified
	}
}
