// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"net"
	"os"
	"time"

	"heimdall/app/internal/domain"
)

// CollectFunc returns the current privileged metrics.
type CollectFunc func(ctx context.Context) []domain.Metric

// Server serves privileged metrics over a local unix socket. Each connection
// receives a single JSON envelope of the latest metrics. It is read-only: the
// client sends nothing and cannot influence what is collected, so the helper
// presents no argument-injection or command-execution surface to callers.
type Server struct {
	SockPath string
	Collect  CollectFunc
	Timeout  time.Duration
}

// Serve listens until ctx is cancelled. It removes any stale socket first and
// loosens the socket permissions so the unprivileged daemon can connect to the
// root-owned helper (the payload is low-sensitivity, read-only telemetry).
func (s *Server) Serve(ctx context.Context) error {
	if s.SockPath == "" {
		s.SockPath = DefaultSocketPath()
	}
	_ = os.Remove(s.SockPath)
	lis, err := net.Listen("unix", s.SockPath)
	if err != nil {
		return err
	}
	_ = os.Chmod(s.SockPath, 0o666)

	go func() {
		<-ctx.Done()
		_ = lis.Close()
		_ = os.Remove(s.SockPath)
	}()

	for {
		conn, err := lis.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}
		go s.handle(ctx, conn)
	}
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	to := s.Timeout
	if to == 0 {
		to = 3 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()
	_ = conn.SetWriteDeadline(time.Now().Add(to))
	_ = encodeMetrics(conn, s.Collect(cctx))
}
