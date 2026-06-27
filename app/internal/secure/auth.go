// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package secure

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Authenticator authorizes an incoming RPC from its context metadata. It is the
// single server-side authorization seam shared by the hub and the daemon's
// control plane, so an unauthenticated caller cannot reach any handler.
type Authenticator interface {
	Authorize(ctx context.Context) error
}

// NewAuthenticator returns a constant-time bearer-token authenticator, or an
// allow-all authenticator when token is empty (unauthenticated dev mode).
func NewAuthenticator(token string) Authenticator {
	if token == "" {
		return allowAll{}
	}
	return tokenAuth{token: token}
}

type allowAll struct{}

func (allowAll) Authorize(context.Context) error { return nil }

type tokenAuth struct{ token string }

func (a tokenAuth) Authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing token")
	}
	vals := md.Get(TokenMetadataKey)
	if len(vals) == 0 || subtle.ConstantTimeCompare([]byte(vals[0]), []byte(a.token)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid or missing token")
	}
	return nil
}

// UnaryServerInterceptor enforces authorization before any unary handler runs.
func UnaryServerInterceptor(auth Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := auth.Authorize(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor enforces authorization before any streaming handler runs.
func StreamServerInterceptor(auth Authenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := auth.Authorize(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}
