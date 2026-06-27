// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build !windows

package helper

import (
	"context"

	"heimdall/app/internal/domain"
)

// windowsPrivileged is a no-op off Windows; WMI thermal zones are Windows-only.
func windowsPrivileged(context.Context) []domain.Metric { return nil }
