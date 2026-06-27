// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package transport maps between the domain types and the versioned gRPC wire
// types (monitoring.v1). Generated proto types are transport DTOs — the domain
// never imports them; mapping happens only here, at the edge.
package transport

import (
	"time"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

func statusToProto(s domain.MetricStatus) v1.MetricStatus {
	switch s {
	case domain.StatusOK:
		return v1.MetricStatus_METRIC_STATUS_OK
	case domain.StatusUnavailable:
		return v1.MetricStatus_METRIC_STATUS_UNAVAILABLE
	case domain.StatusInsufficientPermission:
		return v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION
	case domain.StatusError:
		return v1.MetricStatus_METRIC_STATUS_ERROR
	default:
		return v1.MetricStatus_METRIC_STATUS_UNSPECIFIED
	}
}

func statusFromProto(s v1.MetricStatus) domain.MetricStatus {
	switch s {
	case v1.MetricStatus_METRIC_STATUS_OK:
		return domain.StatusOK
	case v1.MetricStatus_METRIC_STATUS_UNAVAILABLE:
		return domain.StatusUnavailable
	case v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION:
		return domain.StatusInsufficientPermission
	case v1.MetricStatus_METRIC_STATUS_ERROR:
		return domain.StatusError
	default:
		return domain.StatusUnspecified
	}
}

// MetricToProto maps a domain metric to a wire sample. A value is sent only for
// OK metrics; non-OK metrics carry status + detail and no value. Per-core
// metrics ride the per_core oneof arm; everything else is a gauge.
func MetricToProto(m domain.Metric) *v1.MetricSample {
	s := &v1.MetricSample{
		Metric: m.Name,
		Status: statusToProto(m.Status),
		Unit:   m.Unit,
		Detail: m.Detail,
	}
	if m.Status == domain.StatusOK {
		if m.Kind == domain.KindPerCore {
			s.Value = &v1.MetricSample_PerCore{PerCore: &v1.PerCore{Values: m.PerCore}}
		} else {
			s.Value = &v1.MetricSample_Gauge{Gauge: m.Gauge}
		}
	}
	return s
}

// MetricFromProto maps a wire sample back to a domain metric.
func MetricFromProto(s *v1.MetricSample) domain.Metric {
	m := domain.Metric{
		Name:   s.GetMetric(),
		Status: statusFromProto(s.GetStatus()),
		Unit:   s.GetUnit(),
		Detail: s.GetDetail(),
	}
	if pc := s.GetPerCore(); pc != nil {
		m.Kind = domain.KindPerCore
		m.PerCore = pc.GetValues()
	} else {
		m.Gauge = s.GetGauge()
	}
	return m
}

// ToSnapshot builds a keyframe snapshot for a host's metrics.
func ToSnapshot(hostID string, ms []domain.Metric, seq uint64, ts time.Time) *v1.Snapshot {
	samples := make([]*v1.MetricSample, 0, len(ms))
	for _, m := range ms {
		samples = append(samples, MetricToProto(m))
	}
	return &v1.Snapshot{
		HostId:       hostID,
		TsUnixMillis: ts.UnixMilli(),
		Seq:          seq,
		Keyframe:     true,
		Samples:      samples,
	}
}

// FromSnapshot extracts the host id and metrics from a wire snapshot.
func FromSnapshot(s *v1.Snapshot) (string, []domain.Metric) {
	ms := make([]domain.Metric, 0, len(s.GetSamples()))
	for _, x := range s.GetSamples() {
		ms = append(ms, MetricFromProto(x))
	}
	return s.GetHostId(), ms
}
