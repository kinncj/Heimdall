// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package discovery

import (
	"context"
	"errors"
	"testing"
)

func TestStaticReturnsAddrOrNotFound(t *testing.T) {
	if a, err := (Static{Addr: "hub:9090"}).Discover(context.Background()); err != nil || a != "hub:9090" {
		t.Fatalf("static = %q, %v; want hub:9090", a, err)
	}
	if _, err := (Static{}).Discover(context.Background()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty static should be ErrNotFound, got %v", err)
	}
}

type stub struct {
	addr string
	err  error
}

func (s stub) Discover(context.Context) (string, error) { return s.addr, s.err }

func TestChainReturnsFirstHit(t *testing.T) {
	c := Chain{
		stub{err: ErrNotFound},
		nil,
		Static{Addr: "second:9090"},
		Static{Addr: "third:9090"},
	}
	a, err := c.Discover(context.Background())
	if err != nil || a != "second:9090" {
		t.Fatalf("chain = %q, %v; want second:9090", a, err)
	}
}

func TestChainAllMissYieldsNotFound(t *testing.T) {
	c := Chain{stub{err: ErrNotFound}, Static{}}
	if _, err := c.Discover(context.Background()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
