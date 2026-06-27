// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package discovery (Ratatoskr) lets a daemon find its hub without a hardcoded
// address. A Discoverer resolves only WHERE the hub is; the enrollment token and
// TLS still gate WHETHER to trust it, so discovery never widens the trust model.
// See ADR 0009.
package discovery

import (
	"context"
	"errors"
)

// ErrNotFound means no hub was discovered.
var ErrNotFound = errors.New("discovery: no hub found")

// Discoverer resolves a hub address ("host:port").
type Discoverer interface {
	Discover(ctx context.Context) (string, error)
}

// Static returns a fixed seed address. Empty yields ErrNotFound so it composes
// cleanly in a Chain as a fallback.
type Static struct{ Addr string }

// Discover returns the static address.
func (s Static) Discover(context.Context) (string, error) {
	if s.Addr == "" {
		return "", ErrNotFound
	}
	return s.Addr, nil
}

// Chain tries each discoverer in order and returns the first hit — e.g. mDNS on
// the LAN, then a static seed for overlay networks. Nil entries are skipped.
type Chain []Discoverer

// Discover returns the first non-empty address any member resolves.
func (c Chain) Discover(ctx context.Context) (string, error) {
	for _, d := range c {
		if d == nil {
			continue
		}
		if addr, err := d.Discover(ctx); err == nil && addr != "" {
			return addr, nil
		}
	}
	return "", ErrNotFound
}
