// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build !linux

package helper

import (
	"context"

	"heimdall/app/internal/domain"
)

// linuxPrivileged is a no-op off Linux; RAPL and hwmon are Linux-only.
func linuxPrivileged(context.Context) []domain.Metric { return nil }
