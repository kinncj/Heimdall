// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package secure builds the gRPC transport credentials shared by the hub
// (server) and the daemon and dashboard (clients). TLS is opt-in: with no
// certificate configured the channel stays plaintext for local development,
// matching the daemon's existing insecure default.
package secure

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TokenMetadataKey is the gRPC metadata key carrying the enrollment token on
// every client RPC.
const TokenMetadataKey = "x-heimdall-token"

// TokenCredentials is a client-side PerRPCCredentials that attaches the
// enrollment token to every outbound RPC. RequireTransportSecurity is false so
// it also works on the plaintext dev channel; over TLS the token rides the
// encrypted connection unchanged.
type TokenCredentials struct{ Token string }

func (c TokenCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{TokenMetadataKey: c.Token}, nil
}

func (TokenCredentials) RequireTransportSecurity() bool { return false }

// ServerOption returns the gRPC server credentials option for the given
// certificate and key. When both are empty it returns a no-op option, leaving
// the server plaintext.
func ServerOption(certFile, keyFile string) (grpc.ServerOption, error) {
	if certFile == "" && keyFile == "" {
		return grpc.EmptyServerOption{}, nil
	}
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("secure: both --tls-cert and --tls-key are required for TLS")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("secure: load key pair: %w", err)
	}
	return grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	})), nil
}

// ClientConfig configures client-side transport security.
type ClientConfig struct {
	Enabled    bool   // when false, the channel is plaintext
	CAFile     string // PEM bundle to trust; empty uses the system roots
	ServerName string // overrides the SNI / verified name
	SkipVerify bool   // dev only: accept any server certificate
}

// DialOption returns the gRPC dial option implementing cfg.
func (cfg ClientConfig) DialOption() (grpc.DialOption, error) {
	if !cfg.Enabled {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}
	tc := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: cfg.ServerName}
	if cfg.SkipVerify {
		tc.InsecureSkipVerify = true
	} else if cfg.CAFile != "" {
		pem, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("secure: read CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("secure: no certificates parsed from %s", cfg.CAFile)
		}
		tc.RootCAs = pool
	}
	return grpc.WithTransportCredentials(credentials.NewTLS(tc)), nil
}
