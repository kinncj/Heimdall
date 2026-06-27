// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"heimdall/app/internal/secure"
	v1 "heimdall/common/proto/monitoring/v1"
)

func serve(t *testing.T, token string) (*bufconn.Listener, *Hub, func()) {
	t.Helper()
	h := New(2*time.Second, 5*time.Second)
	if token != "" {
		h.SetToken(token)
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(h.UnaryInterceptor()),
		grpc.StreamInterceptor(h.StreamInterceptor()),
	)
	v1.RegisterEnrollmentServiceServer(srv, h)
	v1.RegisterMetricStreamServiceServer(srv, h)
	v1.RegisterFederationServiceServer(srv, h)

	lis := bufconn.Listen(1 << 20)
	go func() { _ = srv.Serve(lis) }()
	return lis, h, func() { srv.Stop(); _ = lis.Close() }
}

func dial(t *testing.T, lis *bufconn.Listener, token string) *grpc.ClientConn {
	t.Helper()
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(secure.TokenCredentials{Token: token}))
	}
	conn, err := grpc.NewClient("passthrough:///bufnet", opts...)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func enroll(conn *grpc.ClientConn) error {
	_, err := v1.NewEnrollmentServiceClient(conn).Enroll(context.Background(), &v1.EnrollRequest{
		Host: &v1.Host{HostId: "alpha", Hostname: "alpha", DisplayName: "alpha"},
	})
	return err
}

// Covers story 0001: enrollment requires a valid token, and a reconnecting
// daemon re-presenting the same HostID is matched to the existing host rather
// than duplicated.
func TestEnrollmentTokenAndStableHostID(t *testing.T) {
	lis, h, stop := serve(t, "s3cret")
	defer stop()

	noTok := dial(t, lis, "")
	defer noTok.Close()
	if st, _ := status.FromError(enroll(noTok)); st.Code() != codes.Unauthenticated {
		t.Fatalf("missing token: got %v, want Unauthenticated", st.Code())
	}

	badTok := dial(t, lis, "nope")
	defer badTok.Close()
	if st, _ := status.FromError(enroll(badTok)); st.Code() != codes.Unauthenticated {
		t.Fatalf("wrong token: got %v, want Unauthenticated", st.Code())
	}

	if got := h.Registry().Count(); got != 0 {
		t.Fatalf("unauthenticated daemon registered: count=%d, want 0", got)
	}

	okTok := dial(t, lis, "s3cret")
	defer okTok.Close()
	if err := enroll(okTok); err != nil {
		t.Fatalf("valid token enroll: %v", err)
	}
	if err := enroll(okTok); err != nil { // reconnect, same HostID
		t.Fatalf("re-enroll: %v", err)
	}
	if got := h.Registry().Count(); got != 1 {
		t.Fatalf("reconnect duplicated host: count=%d, want 1", got)
	}
}

// A daemon cannot bypass enrollment by opening the metric stream directly: the
// stream interceptor enforces the same token.
func TestStreamRequiresToken(t *testing.T) {
	lis, _, stop := serve(t, "s3cret")
	defer stop()

	conn := dial(t, lis, "")
	defer conn.Close()

	stream, err := v1.NewMetricStreamServiceClient(conn).Stream(context.Background())
	if err != nil {
		if st, _ := status.FromError(err); st.Code() == codes.Unauthenticated {
			return
		}
		t.Fatalf("open stream: %v", err)
	}
	_ = stream.Send(&v1.Snapshot{HostId: "alpha"})
	_, recvErr := stream.Recv()
	if st, _ := status.FromError(recvErr); st.Code() != codes.Unauthenticated {
		t.Fatalf("unauthenticated stream: got %v, want Unauthenticated", st.Code())
	}
}
