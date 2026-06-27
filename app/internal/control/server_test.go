// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package control

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	v1 "heimdall/common/proto/monitoring/v1"
)

func TestControlServiceOverWire(t *testing.T) {
	var audit bytes.Buffer
	exec := &Executor{
		Allow: Allowlist{"echo.test": {Key: "echo.test", Argv: []string{"echo", "ok"}}},
		Audit: NewWriterAuditor(&audit),
	}
	srv := grpc.NewServer()
	v1.RegisterControlPlaneServiceServer(srv, NewServer(exec, "alpha"))
	lis := bufconn.Listen(1 << 20)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := Client{Conn: conn}

	resp, err := client.Run(context.Background(), "alpha", "echo.test", nil, "op@dash")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if resp.GetStatus() != v1.MetricStatus_METRIC_STATUS_OK || strings.TrimSpace(resp.GetStdout()) != "ok" {
		t.Fatalf("allow-listed response = %+v", resp)
	}

	refused, err := client.Run(context.Background(), "alpha", "rm.rf", []string{"/"}, "mallory")
	if err != nil {
		t.Fatalf("run refused: %v", err)
	}
	if refused.GetStatus() != v1.MetricStatus_METRIC_STATUS_INSUFFICIENT_PERMISSION {
		t.Fatalf("refused status = %v, want insufficient_permission", refused.GetStatus())
	}
	if refused.GetStdout() != "" {
		t.Errorf("refused command produced output: %q", refused.GetStdout())
	}
	if !strings.Contains(audit.String(), "decision=ALLOW") || !strings.Contains(audit.String(), "decision=REFUSE") {
		t.Errorf("audit log incomplete: %q", audit.String())
	}
}
