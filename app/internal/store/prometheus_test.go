// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package store

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/snappy"
	"google.golang.org/protobuf/proto"

	"heimdall/app/internal/domain"
	promv1 "heimdall/common/proto/prometheus/v1"
)

func TestWriteSendsRemoteWriteSeries(t *testing.T) {
	got := make(chan *promv1.WriteRequest, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "snappy" {
			t.Errorf("missing snappy encoding")
		}
		body, _ := io.ReadAll(r.Body)
		raw, err := snappy.Decode(nil, body)
		if err != nil {
			t.Fatalf("snappy decode: %v", err)
		}
		var wr promv1.WriteRequest
		if err := proto.Unmarshal(raw, &wr); err != nil {
			t.Fatalf("proto unmarshal: %v", err)
		}
		got <- &wr
	}))
	defer srv.Close()

	st := NewPrometheus(srv.URL)
	views := []domain.HostView{{
		Host:         domain.Host{ID: "alpha", Context: domain.HostContext{Labels: map[string]string{"hub": "home"}}},
		State:        domain.StateOnline,
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: 42}},
	}}
	if err := st.Write(context.Background(), views, time.Unix(1_700_000_000, 0)); err != nil {
		t.Fatal(err)
	}

	wr := <-got
	var foundCPU bool
	for _, ts := range wr.Timeseries {
		name, host := "", ""
		for _, l := range ts.Labels {
			switch l.Name {
			case "__name__":
				name = l.Value
			case "host":
				host = l.Value
			}
		}
		if name == "heimdall_cpu_util" && host == "alpha" && ts.Samples[0].Value == 42 {
			foundCPU = true
		}
	}
	if !foundCPU {
		t.Errorf("remote-write missing heimdall_cpu_util{host=alpha} 42: %+v", wr.Timeseries)
	}
}

func TestRestoreReconstructsHosts(t *testing.T) {
	const body = `{"status":"success","data":{"resultType":"vector","result":[
	  {"metric":{"__name__":"heimdall_host_up","host":"alpha","hub":"home","state":"online"},"value":[1700000000,"1"]},
	  {"metric":{"__name__":"heimdall_cpu_util","host":"alpha","hub":"home"},"value":[1700000000,"42"]},
	  {"metric":{"__name__":"heimdall_host_up","host":"beta","hub":"remote","state":"offline"},"value":[1699999000,"1"]}
	]}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	views, err := NewPrometheus(srv.URL).Restore(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 2 {
		t.Fatalf("got %d hosts, want 2 (incl the offline one)", len(views))
	}
	byID := map[domain.HostID]domain.HostView{}
	for _, v := range views {
		byID[v.Host.ID] = v
	}
	alpha, ok := byID["alpha"]
	if !ok || alpha.Host.Context.Labels["hub"] != "home" {
		t.Fatalf("alpha not restored with hub=home: %+v", alpha)
	}
	var cpu float64 = -1
	for _, m := range alpha.LastSnapshot {
		if m.Name == "cpu.util" {
			cpu = m.Gauge
		}
	}
	if cpu != 42 {
		t.Errorf("alpha cpu.util = %v, want 42", cpu)
	}
	// beta is offline (host_up only, no metrics) but must still be restored.
	if _, ok := byID["beta"]; !ok {
		t.Error("offline host beta should be restored from host_up")
	}
}
