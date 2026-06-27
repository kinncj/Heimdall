// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package control

import (
	"errors"
	"io"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

// Server implements the gRPC ControlPlaneService by delegating every request to
// an Executor. It is served by the daemon (which runs as the unprivileged user),
// so commands never escalate.
type Server struct {
	v1.UnimplementedControlPlaneServiceServer
	exec   *Executor
	hostID string
}

// NewServer builds a control server for the given executor and host id.
func NewServer(exec *Executor, hostID string) *Server {
	return &Server{exec: exec, hostID: hostID}
}

// Execute runs each requested command and streams back its bounded result.
func (s *Server) Execute(stream v1.ControlPlaneService_ExecuteServer) error {
	ctx := stream.Context()
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		res := s.exec.Execute(ctx, req.GetAllowlistedCmd(), req.GetArgs(), req.GetActor())
		if err := stream.Send(&v1.ControlResponse{
			RequestId: req.GetRequestId(),
			ExitCode:  int32(res.ExitCode),
			Stdout:    res.Stdout,
			Stderr:    res.Stderr,
			Truncated: res.Truncated,
			Status:    statusToProto(res.Status),
		}); err != nil {
			return err
		}
	}
}

func statusToProto(s domain.MetricStatus) v1.MetricStatus {
	switch s {
	case domain.StatusOK:
		return v1.MetricStatus_METRIC_STATUS_OK
	case domain.StatusInsufficientPermission:
		return v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION
	case domain.StatusError:
		return v1.MetricStatus_METRIC_STATUS_ERROR
	default:
		return v1.MetricStatus_METRIC_STATUS_UNSPECIFIED
	}
}
