// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package discovery

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// Service is the mDNS service type hubs advertise and daemons browse for.
const Service = "_heimdall-hub._tcp"

// MDNS discovers a hub by browsing the local network for the Heimdall service.
// It works on a flat L2 LAN; overlay networks (Tailscale, etc.) have no
// multicast, so pair it with a Static seed in a Chain.
type MDNS struct {
	Service string
	Domain  string
	Timeout time.Duration
}

// Discover returns the first advertised hub's "host:port".
func (m MDNS) Discover(ctx context.Context) (string, error) {
	service, domain := m.Service, m.Domain
	if service == "" {
		service = Service
	}
	if domain == "" {
		domain = "local."
	}
	timeout := m.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return "", err
	}
	entries := make(chan *zeroconf.ServiceEntry, 4)
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := resolver.Browse(cctx, service, domain, entries); err != nil {
		return "", err
	}
	for {
		select {
		case e, ok := <-entries:
			if !ok {
				return "", ErrNotFound
			}
			if e == nil || e.Port == 0 {
				continue
			}
			host := ""
			switch {
			case len(e.AddrIPv4) > 0:
				host = e.AddrIPv4[0].String()
			case e.HostName != "":
				host = strings.TrimSuffix(e.HostName, ".")
			}
			if host == "" {
				continue
			}
			return fmt.Sprintf("%s:%d", host, e.Port), nil
		case <-cctx.Done():
			return "", ErrNotFound
		}
	}
}

// Advertise publishes a hub service over mDNS so discoverable daemons can find
// it. Close the returned closer to stop advertising.
func Advertise(instance string, port int, txt []string) (io.Closer, error) {
	server, err := zeroconf.Register(instance, Service, "local.", port, txt, nil)
	if err != nil {
		return nil, err
	}
	return closerFunc(func() error { server.Shutdown(); return nil }), nil
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }
