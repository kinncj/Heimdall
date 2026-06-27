// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import (
	"google.golang.org/grpc"

	"heimdall/app/internal/secure"
)

// UnaryInterceptor enforces authorization before any unary handler runs.
func (h *Hub) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return secure.UnaryServerInterceptor(h.auth)
}

// StreamInterceptor enforces authorization before any streaming handler runs.
func (h *Hub) StreamInterceptor() grpc.StreamServerInterceptor {
	return secure.StreamServerInterceptor(h.auth)
}
