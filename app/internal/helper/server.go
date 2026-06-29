// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"time"

	"heimdall/app/internal/command"
	"heimdall/app/internal/domain"
)

// CollectFunc returns the current privileged metrics.
type CollectFunc func(ctx context.Context) []domain.Metric

// Server serves privileged metrics — and, opt-in, privileged allow-listed
// commands — over a local unix socket. For commands the helper enforces its OWN
// allow-list (command.IsPrivileged): it never trusts the daemon and will only run
// commands that are both allow-listed and explicitly privileged, with bounded
// output and no shell. A request for anything else is refused.
type Server struct {
	SockPath string
	Collect  CollectFunc
	Timeout  time.Duration
}

// Serve listens until ctx is cancelled. It removes any stale socket first and
// restricts the socket to owner+group (0660). The helper now runs privileged
// allow-listed commands, so a world-writable socket would let ANY local user ask
// root to run them — a local privilege-escalation vector. The daemon reaches the
// root-owned helper via a shared group (see DefaultSocketPath / --socket).
func (s *Server) Serve(ctx context.Context) error {
	if s.SockPath == "" {
		s.SockPath = DefaultSocketPath()
	}
	_ = os.Remove(s.SockPath)
	lis, err := net.Listen("unix", s.SockPath)
	if err != nil {
		return err
	}
	_ = os.Chmod(s.SockPath, 0o660)

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
		// Match the client's Exec timeout (10s) so a privileged command that runs
		// longer than a few seconds isn't cut off server-side while the client waits.
		to = 10 * time.Second
	}
	// Read the request first. An old (silent) client sends nothing, so a short read
	// deadline falls back to "collect" — preserving the metric path.
	_ = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	var req request
	reqErr := json.NewDecoder(conn).Decode(&req)

	cctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()
	_ = conn.SetWriteDeadline(time.Now().Add(to))

	if reqErr == nil && req.Op == "exec" {
		// The helper trusts no caller: only its own privileged allow-list runs.
		if !command.IsPrivileged(req.Cmd) {
			_ = encodeResult(conn, command.Result{
				Status: domain.StatusInsufficientPermission, ExitCode: -1,
				Stderr: "command is not a privileged allow-listed command",
			})
			return
		}
		cmdCtx, cmdCancel := context.WithTimeout(ctx, to)
		defer cmdCancel()
		_ = encodeResult(conn, command.Run(cmdCtx, req.Cmd, req.Args))
		return
	}
	_ = encodeMetrics(conn, s.Collect(cctx))
}
