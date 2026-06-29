// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package transport

import (
	"testing"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

func TestCommandResultFromSnapshot(t *testing.T) {
	snap := &v1.Snapshot{CommandResult: &v1.ControlResponse{
		RequestId: "r1", ExitCode: 3, Stdout: "out", Stderr: "err",
		Truncated: true, Status: v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION,
	}}
	cr := CommandResultFromSnapshot(snap)
	if cr == nil {
		t.Fatal("expected a command result")
	}
	if cr.RequestID != "r1" || cr.ExitCode != 3 || cr.Stdout != "out" || cr.Stderr != "err" || !cr.Truncated {
		t.Fatalf("round-trip lost fields: %+v", cr)
	}
	if cr.Status != domain.StatusInsufficientPermission {
		t.Fatalf("status = %v, want insufficient_permission", cr.Status)
	}

	if CommandResultFromSnapshot(&v1.Snapshot{}) != nil {
		t.Fatal("a snapshot without a command result must yield nil")
	}
}

func TestStatusToProtoRoundTrip(t *testing.T) {
	for _, s := range []domain.MetricStatus{
		domain.StatusOK, domain.StatusUnavailable,
		domain.StatusInsufficientPermission, domain.StatusError,
	} {
		if got := statusFromProto(StatusToProto(s)); got != s {
			t.Fatalf("status %v did not round-trip (got %v)", s, got)
		}
	}
}
