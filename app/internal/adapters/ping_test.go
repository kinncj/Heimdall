// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func TestParsePingRTT(t *testing.T) {
	cases := []struct {
		out  string
		want float64
		ok   bool
	}{
		{"64 bytes from 1.1.1.1: icmp_seq=0 ttl=55 time=4.513 ms", 4.513, true},
		{"64 bytes from 192.168.1.1: icmp_seq=1 ttl=64 time<1 ms", 1, true},
		{"Request timeout for icmp_seq 0", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := parsePingRTT(c.out)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("parsePingRTT(%q) = %v,%v; want %v,%v", c.out, got, ok, c.want, c.ok)
		}
	}
}

func TestPingArgsByOS(t *testing.T) {
	a := pingArgs("1.1.1.1", 2*time.Second)
	if a[len(a)-1] != "1.1.1.1" || a[0] != "-c" || a[1] != "1" {
		t.Fatalf("pingArgs = %v", a)
	}
}

func TestReachabilityUsesPing(t *testing.T) {
	orig := pingFn
	defer func() { pingFn = orig }()

	pingFn = func(context.Context, string, time.Duration) (float64, bool) { return 12.5, true }
	ms := first(t, Reachability{Target: "9.9.9.9"})
	if ms.Name != "net.latency" || ms.Status != domain.StatusOK || ms.Gauge != 12.5 {
		t.Fatalf("reachable = %+v", ms)
	}

	pingFn = func(context.Context, string, time.Duration) (float64, bool) { return 0, false }
	down := first(t, Reachability{Target: "9.9.9.9"})
	if down.Status != domain.StatusError {
		t.Fatalf("unreachable = %+v, want error", down)
	}
}

func TestProcNetRouteParsing(t *testing.T) {
	// Two NICs with default routes: eth0 -> 192.168.1.1, wlan0 -> 10.0.0.1.
	data := "Iface\tDestination\tGateway\tFlags\tRefCnt\tUse\tMetric\tMask\n" +
		"eth0\t00000000\t0101A8C0\t0003\t0\t0\t100\t00000000\t0\t0\t0\n" +
		"eth0\t0001A8C0\t00000000\t0001\t0\t0\t100\t00FFFFFF\t0\t0\t0\n" +
		"wlan0\t00000000\t0100000A\t0003\t0\t0\t600\t00000000\t0\t0\t0\n"
	gws := parseProcNetRoute(data)
	if len(gws) != 2 {
		t.Fatalf("got %d gateways, want 2: %+v", len(gws), gws)
	}
	byIface := map[string]string{}
	for _, g := range gws {
		byIface[g.iface] = g.ip
	}
	if byIface["eth0"] != "192.168.1.1" {
		t.Errorf("eth0 gw = %q, want 192.168.1.1", byIface["eth0"])
	}
	if byIface["wlan0"] != "10.0.0.1" {
		t.Errorf("wlan0 gw = %q, want 10.0.0.1", byIface["wlan0"])
	}
}

func TestNetstatRouteParsing(t *testing.T) {
	// macOS netstat -rn -f inet: two default routes (en0, en8) plus a link route.
	data := `Routing tables

Internet:
Destination        Gateway            Flags        Netif Expire
default            192.168.1.1        UGScg          en0
default            10.0.0.1           UGScgI         en8
127                127.0.0.1          UCS            lo0
192.168.1          link#14            UCS            en0`
	gws := parseNetstatRoutes(data)
	if len(gws) != 2 {
		t.Fatalf("got %d gateways, want 2: %+v", len(gws), gws)
	}
	byIface := map[string]string{}
	for _, g := range gws {
		byIface[g.iface] = g.ip
	}
	if byIface["en0"] != "192.168.1.1" || byIface["en8"] != "10.0.0.1" {
		t.Errorf("gateways = %+v", byIface)
	}
}

func TestGatewayAdapterPingsEachNIC(t *testing.T) {
	orig := pingFn
	defer func() { pingFn = orig }()
	pingFn = func(_ context.Context, ip string, _ time.Duration) (float64, bool) {
		switch ip {
		case "192.168.1.1":
			return 2.0, true
		case "10.0.0.1":
			return 9.0, true
		}
		return 0, false
	}

	g := Gateway{discover: func() []gateway {
		return []gateway{{iface: "en0", ip: "192.168.1.1"}, {iface: "en8", ip: "10.0.0.1"}}
	}}
	ms, err := g.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]domain.Metric{}
	for _, m := range ms {
		got[m.Name] = m
	}
	if got["net.gateway.en0"].Gauge != 2.0 || got["net.gateway.en8"].Gauge != 9.0 {
		t.Errorf("per-NIC = %+v", got)
	}
	// Primary is the lowest-latency reachable gateway.
	if got["net.gateway"].Status != domain.StatusOK || got["net.gateway"].Gauge != 2.0 {
		t.Errorf("primary net.gateway = %+v, want 2.0ms", got["net.gateway"])
	}
}

func TestGatewayAdapterNoRoute(t *testing.T) {
	g := Gateway{discover: func() []gateway { return nil }}
	m := first(t, g)
	if m.Name != "net.gateway" || m.Status != domain.StatusUnavailable {
		t.Fatalf("no-route = %+v, want net.gateway unavailable", m)
	}
}
