// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import "strings"

// Address is a network endpoint such as "localhost:9090". The zero value is
// empty, meaning "no endpoint".
type Address struct{ raw string }

// NewAddress trims and wraps a raw address string.
func NewAddress(raw string) Address { return Address{raw: strings.TrimSpace(raw)} }

func (a Address) String() string { return a.raw }
func (a Address) IsEmpty() bool  { return a.raw == "" }

// IsLoopback reports whether the host part is a loopback address.
func (a Address) IsLoopback() bool {
	host := a.raw
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	host = strings.Trim(host, "[]")
	return host == "" || host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// Secret is a sensitive value (a token) that never reveals itself when printed.
type Secret struct{ raw string }

// NewSecret wraps a raw secret string.
func NewSecret(raw string) Secret { return Secret{raw: raw} }

// Reveal returns the underlying secret, for use at the trust boundary only.
func (s Secret) Reveal() string { return s.raw }
func (s Secret) IsEmpty() bool  { return s.raw == "" }

// String redacts the secret so it is safe to log.
func (s Secret) String() string {
	if s.raw == "" {
		return ""
	}
	return "••••••"
}
