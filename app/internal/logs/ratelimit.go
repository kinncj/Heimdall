// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package logs

import (
	"sync"
	"time"
)

// rateLimiter caps events per second with a coarse fixed window: simple,
// allocation-free, and enough to keep log volume within a low-bandwidth budget.
type rateLimiter struct {
	mu       sync.Mutex
	max      int
	windowAt time.Time
	count    int
}

func newRateLimiter(maxPerSec int) *rateLimiter {
	return &rateLimiter{max: maxPerSec}
}

// allow reports whether an event may be sent now, counting it if so. A max <= 0
// disables limiting.
func (r *rateLimiter) allow(now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if now.Sub(r.windowAt) >= time.Second {
		r.windowAt = now
		r.count = 0
	}
	if r.max > 0 && r.count >= r.max {
		return false
	}
	r.count++
	return true
}
