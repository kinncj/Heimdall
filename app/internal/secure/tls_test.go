// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package secure_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"heimdall/app/internal/hub"
	"heimdall/app/internal/secure"
	v1 "heimdall/common/proto/monitoring/v1"
)

func genCert(t *testing.T, dnsName string) (certFile, keyFile string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: dnsName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{dnsName},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	writePEM(t, certFile, "CERTIFICATE", der)
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	writePEM(t, keyFile, "EC PRIVATE KEY", keyBytes)
	return certFile, keyFile
}

func writePEM(t *testing.T, path, typ string, der []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: typ, Bytes: der}); err != nil {
		t.Fatal(err)
	}
}

// Proves the secure package's server and client credentials interoperate: a TLS
// client trusting the self-signed cert enrolls, while a plaintext client against
// the same TLS server is rejected at the transport layer.
func TestTLSEnrollRoundTrip(t *testing.T) {
	const dns = "heimdall-hub"
	certFile, keyFile := genCert(t, dns)

	h := hub.New(2*time.Second, 5*time.Second)
	h.SetToken("s3cret")
	srvOpt, err := secure.ServerOption(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer(srvOpt,
		grpc.UnaryInterceptor(h.UnaryInterceptor()),
		grpc.StreamInterceptor(h.StreamInterceptor()),
	)
	v1.RegisterEnrollmentServiceServer(srv, h)
	lis := bufconn.Listen(1 << 20)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	dialer := grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	})
	tok := grpc.WithPerRPCCredentials(secure.TokenCredentials{Token: "s3cret"})

	do, err := secure.ClientConfig{Enabled: true, CAFile: certFile, ServerName: dns}.DialOption()
	if err != nil {
		t.Fatal(err)
	}
	conn, err := grpc.NewClient("passthrough:///bufnet", dialer, do, tok)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := v1.NewEnrollmentServiceClient(conn).Enroll(context.Background(),
		&v1.EnrollRequest{Host: &v1.Host{HostId: "alpha"}}); err != nil {
		t.Fatalf("tls enroll: %v", err)
	}

	plain, err := secure.ClientConfig{Enabled: false}.DialOption()
	if err != nil {
		t.Fatal(err)
	}
	pconn, err := grpc.NewClient("passthrough:///bufnet", dialer, plain, tok)
	if err != nil {
		t.Fatal(err)
	}
	defer pconn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := v1.NewEnrollmentServiceClient(pconn).Enroll(ctx,
		&v1.EnrollRequest{Host: &v1.Host{HostId: "beta"}}); err == nil {
		t.Fatal("plaintext client unexpectedly enrolled against TLS server")
	}
}
