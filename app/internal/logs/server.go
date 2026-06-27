// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package logs

import (
	"strings"
	"time"

	v1 "heimdall/common/proto/monitoring/v1"
)

// serverMaxLinesPerSec is the hard cap the server enforces regardless of what a
// client requests, protecting the link from a noisy source.
const serverMaxLinesPerSec = 200

// Server implements LogStreamService.Tail over a daemon's opt-in sources. It is
// served by the daemon on the same authenticated endpoint as the control plane,
// but on a distinct gRPC service, keeping logs independent of the metric stream.
type Server struct {
	v1.UnimplementedLogStreamServiceServer
	sources Sources
	hostID  string
}

// NewServer builds a log server over the given opt-in source registry.
func NewServer(sources Sources, hostID string) *Server {
	return &Server{sources: sources, hostID: hostID}
}

// Tail streams new lines from the requested opt-in sources, rate-limited, until
// the client disconnects. Unknown or unconfigured sources yield nothing, so log
// streaming stays off until a source is explicitly configured.
func (s *Server) Tail(req *v1.LogTailRequest, stream v1.LogStreamService_TailServer) error {
	ctx := stream.Context()

	type src struct{ alias, path string }
	var srcs []src
	for _, alias := range req.GetSources() {
		if p, ok := s.sources.Resolve(alias); ok {
			srcs = append(srcs, src{alias: alias, path: p})
		}
	}
	if len(srcs) == 0 {
		return nil // nothing opted in -> no logs streamed
	}

	max := int(req.GetMaxLinesPerSec())
	if max <= 0 || max > serverMaxLinesPerSec {
		max = serverMaxLinesPerSec
	}
	limiter := newRateLimiter(max)
	filter := req.GetFilter()

	lines := make(chan *v1.LogLine, 256)
	for _, sc := range srcs {
		sc := sc
		go func() {
			_ = tailFile(ctx, sc.path, func(line string) {
				if filter != "" && !strings.Contains(line, filter) {
					return
				}
				select {
				case lines <- &v1.LogLine{
					HostId:       s.hostID,
					Source:       sc.alias,
					TsUnixMillis: time.Now().UnixMilli(),
					Line:         line,
				}:
				case <-ctx.Done():
				}
			})
		}()
	}

	dropped := false
	for {
		select {
		case <-ctx.Done():
			return nil
		case ll := <-lines:
			if !limiter.allow(time.Now()) {
				dropped = true
				continue
			}
			if dropped {
				ll.RateLimited = true
				dropped = false
			}
			if err := stream.Send(ll); err != nil {
				return err
			}
		}
	}
}
