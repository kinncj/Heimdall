// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"

	"heimdall/app/internal/domain"
)

// first collects from a, asserts it produced at least one metric, and returns
// the first. Shared by the per-adapter smoke tests.
func first(t *testing.T, a domain.Adapter) domain.Metric {
	t.Helper()
	ms, err := a.Collect(context.Background())
	if err != nil {
		t.Fatalf("%s collect: %v", a.Describe().ID, err)
	}
	if len(ms) == 0 {
		t.Fatalf("%s: no metrics", a.Describe().ID)
	}
	return ms[0]
}

func TestDefaultAdapters(t *testing.T) {
	if n := len(Default()); n != 10 {
		t.Fatalf("Default() = %d adapters, want 10", n)
	}
}
