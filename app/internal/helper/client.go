// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"errors"
	"net"
	"time"

	"heimdall/app/internal/command"
	"heimdall/app/internal/domain"
)

// ErrUnavailable means the helper socket could not be reached — typically the
// helper is not installed or not running. Callers degrade to a needs-helper
// affordance rather than treating this as a hard error.
var ErrUnavailable = errors.New("helper: socket unavailable")

// Client reads privileged metrics from the helper's local socket.
type Client struct {
	SockPath string
	Timeout  time.Duration
}

// Collect dials the helper and decodes one metric envelope. A dial failure
// (helper absent) returns ErrUnavailable.
func (c Client) Collect(ctx context.Context) ([]domain.Metric, error) {
	sock := c.SockPath
	if sock == "" {
		sock = DefaultSocketPath()
	}
	to := c.Timeout
	if to == 0 {
		to = 2 * time.Second
	}
	d := net.Dialer{Timeout: to}
	conn, err := d.DialContext(ctx, "unix", sock)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(to))
	if err := encodeRequest(conn, request{V: protocolVersion, Op: "collect"}); err != nil {
		return nil, err
	}
	return decodeMetrics(conn)
}

// Exec asks the root helper to run a privileged allow-listed command (v2 Phase
// 2b). A dial failure (helper absent) returns ErrUnavailable so the daemon can
// report that the command needs the helper. The helper enforces its own allow-list.
func (c Client) Exec(ctx context.Context, cmd string, args []string) (command.Result, error) {
	sock := c.SockPath
	if sock == "" {
		sock = DefaultSocketPath()
	}
	to := c.Timeout
	if to == 0 {
		to = 10 * time.Second // privileged commands may run longer than a metric read
	}
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "unix", sock)
	if err != nil {
		return command.Result{}, ErrUnavailable
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(to))
	if err := encodeRequest(conn, request{V: protocolVersion, Op: "exec", Cmd: cmd, Args: args}); err != nil {
		return command.Result{}, err
	}
	return decodeResult(conn)
}
