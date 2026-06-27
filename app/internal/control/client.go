// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package control

import (
	"context"

	"google.golang.org/grpc"

	v1 "heimdall/common/proto/monitoring/v1"
)

// Client invokes allow-listed control commands against a daemon's control
// endpoint over an existing gRPC connection.
type Client struct {
	Conn *grpc.ClientConn
}

// Run executes one allow-listed command and returns its bounded response.
func (c Client) Run(ctx context.Context, hostID, cmd string, args []string, actor string) (*v1.ControlResponse, error) {
	stream, err := v1.NewControlPlaneServiceClient(c.Conn).Execute(ctx)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&v1.ControlRequest{
		RequestId:      "1",
		HostId:         hostID,
		AllowlistedCmd: cmd,
		Args:           args,
		Actor:          actor,
	}); err != nil {
		return nil, err
	}
	resp, err := stream.Recv()
	_ = stream.CloseSend()
	return resp, err
}
