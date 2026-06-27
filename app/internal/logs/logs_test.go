// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package logs

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	v1 "heimdall/common/proto/monitoring/v1"
)

func TestParseSources(t *testing.T) {
	off, err := ParseSources("")
	if err != nil || len(off) != 0 {
		t.Fatalf("blank spec = %v, %v; want empty registry", off, err)
	}
	s, err := ParseSources("app=/var/log/app.log, sys=/var/log/system.log")
	if err != nil {
		t.Fatal(err)
	}
	if p, ok := s.Resolve("app"); !ok || p != "/var/log/app.log" {
		t.Errorf("resolve app = %q, %v", p, ok)
	}
	if _, ok := s.Resolve("unknown"); ok {
		t.Error("unknown alias resolved")
	}
	if _, err := ParseSources("bad-no-equals"); err == nil {
		t.Error("malformed spec accepted")
	}
}

func TestRateLimiter(t *testing.T) {
	now := time.Now()
	r := newRateLimiter(2)
	if !r.allow(now) || !r.allow(now) {
		t.Fatal("first two events should be allowed")
	}
	if r.allow(now) {
		t.Fatal("third event in the same second should be blocked")
	}
	if !r.allow(now.Add(time.Second)) {
		t.Fatal("event in the next window should be allowed")
	}
}

func TestTailFileEmitsAppendedLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := os.WriteFile(path, []byte("preexisting\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := make(chan string, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = tailFile(ctx, path, func(l string) { got <- l }) }()

	time.Sleep(200 * time.Millisecond) // let the tailer seek to end
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("alpha\nbravo\n")
	f.Close()

	for _, want := range []string{"alpha", "bravo"} {
		select {
		case line := <-got:
			if line != want {
				t.Fatalf("line = %q, want %q", line, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for %q", want)
		}
	}
}

func TestTailServiceStreamsConfiguredSourceOnly(t *testing.T) {
	dir := t.TempDir()
	appLog := filepath.Join(dir, "app.log")
	os.WriteFile(appLog, []byte(""), 0o644)

	srv := grpc.NewServer()
	v1.RegisterLogStreamServiceServer(srv, NewServer(Sources{"app": appLog}, "alpha"))
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

	// Configured source streams appended lines.
	stream, err := v1.NewLogStreamServiceClient(conn).Tail(context.Background(),
		&v1.LogTailRequest{HostId: "alpha", Sources: []string{"app"}})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	f, _ := os.OpenFile(appLog, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("hello-logs\n")
	f.Close()

	line, err := stream.Recv()
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if line.GetLine() != "hello-logs" || line.GetSource() != "app" {
		t.Fatalf("line = %+v", line)
	}

	// Unknown / unconfigured source streams nothing (opt-in off).
	off, err := v1.NewLogStreamServiceClient(conn).Tail(context.Background(),
		&v1.LogTailRequest{HostId: "alpha", Sources: []string{"unknown"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := off.Recv(); err == nil {
		t.Fatal("unknown source streamed a line; want EOF")
	}
}
