// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package observe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

type fakeSource struct{ views []domain.HostView }

func (f fakeSource) Hosts() []domain.HostView { return f.views }

func TestHandlerServesMetrics(t *testing.T) {
	src := fakeSource{views: []domain.HostView{{
		Host:         domain.Host{ID: "alpha"},
		State:        domain.StateOnline,
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: 42}},
	}}}
	srv := httptest.NewServer(Handler(src, nil))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	if !strings.Contains(string(buf[:n]), `heimdall_cpu_util{host="alpha"} 42`) {
		t.Errorf("metrics body missing series:\n%s", buf[:n])
	}
}

func TestHandlerServesHistory(t *testing.T) {
	hist := NewHistory(10)
	hist.Record([]domain.HostView{{
		Host:         domain.Host{ID: "alpha"},
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: 7}},
	}}, time.Unix(1_700_000_000, 0))
	srv := httptest.NewServer(Handler(fakeSource{}, hist))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/history?host=alpha&metric=cpu.util")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got []Sample
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != 7 {
		t.Fatalf("history = %+v, want one sample value 7", got)
	}
}
