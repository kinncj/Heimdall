// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package discovery

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
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

// Found is a discovered hub: its advertised instance name and "host:port".
type Found struct {
	Name string
	Addr string
}

// BrowseAll browses the LAN for the whole timeout and returns every distinct hub
// it sees (deduplicated by address), for choosing among multiple hubs. Order is
// stable (by name then address). It never errors on "none found" — an empty slice
// means no hub was advertised.
func BrowseAll(timeout time.Duration) ([]Found, error) {
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	entries := make(chan *zeroconf.ServiceEntry, 8)
	if err := resolver.Browse(ctx, Service, "local.", entries); err != nil {
		return nil, err
	}

	byAddr := map[string]Found{}
	for {
		select {
		case e, ok := <-entries:
			if !ok {
				return sortedFound(byAddr), nil
			}
			if e == nil || e.Port == 0 {
				continue
			}
			host := ""
			switch {
			case len(e.AddrIPv4) > 0:
				host = e.AddrIPv4[0].String()
			case len(e.AddrIPv6) > 0:
				host = e.AddrIPv6[0].String() // IPv6-only networks: don't fall straight to HostName
			case e.HostName != "":
				host = strings.TrimSuffix(e.HostName, ".")
			}
			if host == "" {
				continue
			}
			// JoinHostPort brackets IPv6 literals ([::1]:9090) so the addr stays dialable.
			addr := net.JoinHostPort(host, strconv.Itoa(int(e.Port)))
			byAddr[addr] = Found{Name: e.Instance, Addr: addr}
		case <-ctx.Done():
			return sortedFound(byAddr), nil
		}
	}
}

func sortedFound(m map[string]Found) []Found {
	out := make([]Found, 0, len(m))
	for _, f := range m {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Addr < out[j].Addr
	})
	return out
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
