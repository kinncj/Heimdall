// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"heimdall/app/internal/domain"
)

// gateway is one default-route next hop and the interface it is reached through.
type gateway struct {
	iface string
	ip    string
}

// Gateway pings the default gateway of every network interface. A machine with
// several NICs has several default routes; each is discovered and probed
// independently. It emits net.gateway for the primary route plus a
// net.gateway.<iface> latency per NIC. discover and the package pingFn are
// injectable for tests.
type Gateway struct {
	discover func() []gateway
}

func (Gateway) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "gateway", Metrics: []string{"net.gateway"}}
}

func (g Gateway) Collect(ctx context.Context) ([]domain.Metric, error) {
	discover := g.discover
	if discover == nil {
		discover = discoverGateways
	}
	gws := discover()
	if len(gws) == 0 {
		return []domain.Metric{{Name: "net.gateway", Status: domain.StatusUnavailable, Detail: "no default route"}}, nil
	}

	type result struct {
		gw gateway
		ms float64
		ok bool
	}
	results := make([]result, len(gws))
	var wg sync.WaitGroup
	for i, gw := range gws {
		wg.Add(1)
		go func(i int, gw gateway) {
			defer wg.Done()
			ms, ok := pingFn(ctx, gw.ip, 1*time.Second)
			results[i] = result{gw: gw, ms: ms, ok: ok}
		}(i, gw)
	}
	wg.Wait()

	out := make([]domain.Metric, 0, len(results)+1)
	var primary *result
	for i := range results {
		r := results[i]
		name := "net.gateway." + sanitizeIface(r.gw.iface)
		if r.ok {
			out = append(out, domain.Metric{Name: name, Unit: "ms", Status: domain.StatusOK, Gauge: r.ms, Detail: r.gw.ip})
			if primary == nil || r.ms < primary.ms {
				primary = &results[i]
			}
		} else {
			out = append(out, domain.Metric{Name: name, Status: domain.StatusError, Detail: "no reply from " + r.gw.ip})
		}
	}

	// net.gateway is the primary (lowest-latency reachable) gateway, or an error
	// when no gateway replied.
	if primary != nil {
		out = append(out, domain.Metric{
			Name: "net.gateway", Unit: "ms", Status: domain.StatusOK, Gauge: primary.ms,
			Detail: fmt.Sprintf("%s via %s", primary.gw.ip, primary.gw.iface),
		})
	} else {
		out = append(out, domain.Metric{Name: "net.gateway", Status: domain.StatusError, Detail: "no gateway replied"})
	}
	return out, nil
}

func sanitizeIface(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

// discoverGateways returns the default-route gateways per interface for the
// current OS.
func discoverGateways() []gateway {
	switch runtime.GOOS {
	case "linux":
		data, err := readProcNetRoute()
		if err != nil {
			return nil
		}
		return parseProcNetRoute(data)
	case "darwin", "freebsd", "openbsd", "netbsd":
		out, err := exec.Command("netstat", "-rn", "-f", "inet").Output()
		if err != nil {
			return nil
		}
		return parseNetstatRoutes(string(out))
	default:
		return nil
	}
}

// parseProcNetRoute parses Linux /proc/net/route, returning the gateway of every
// default route (Destination 00000000). Gateway and destination are
// little-endian hex.
func parseProcNetRoute(data string) []gateway {
	var out []gateway
	for i, line := range strings.Split(data, "\n") {
		if i == 0 { // header
			continue
		}
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		if f[1] != "00000000" { // not a default route
			continue
		}
		ip := hexLEToIP(f[2])
		if ip == "" || ip == "0.0.0.0" {
			continue
		}
		out = append(out, gateway{iface: f[0], ip: ip})
	}
	return dedupeGateways(out)
}

func hexLEToIP(h string) string {
	if len(h) != 8 {
		return ""
	}
	b := make([]byte, 4)
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
		if err != nil {
			return ""
		}
		b[i] = byte(v)
	}
	// /proc/net/route stores the address little-endian, so reverse the bytes.
	return net.IPv4(b[3], b[2], b[1], b[0]).String()
}

func readProcNetRoute() (string, error) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// parseNetstatRoutes parses `netstat -rn -f inet` (macOS/BSD), returning the
// gateway of every default route that has a real IP next hop. link# gateways
// (directly attached routes) are skipped.
func parseNetstatRoutes(data string) []gateway {
	var out []gateway
	for _, line := range strings.Split(data, "\n") {
		f := strings.Fields(line)
		if len(f) < 4 || f[0] != "default" {
			continue
		}
		gwIP := f[1]
		if net.ParseIP(gwIP) == nil { // e.g. "link#14"
			continue
		}
		iface := f[len(f)-1] // Netif is the last column on macOS
		if net.ParseIP(iface) != nil || strings.Contains(iface, ":") {
			iface = f[3] // fall back to the conventional Netif column
		}
		out = append(out, gateway{iface: iface, ip: gwIP})
	}
	return dedupeGateways(out)
}

func dedupeGateways(in []gateway) []gateway {
	seen := make(map[string]bool, len(in))
	var out []gateway
	for _, g := range in {
		key := g.iface + "|" + g.ip
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].iface < out[j].iface })
	return out
}
