// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/snappy"
	"google.golang.org/protobuf/proto"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/observe"
	promv1 "heimdall/common/proto/prometheus/v1"
)

// Prometheus is a Store backed by any Prometheus-compatible TSDB: it pushes via
// remote-write (/api/v1/write) and restores via the instant-query API
// (/api/v1/query). It holds no state of its own.
type Prometheus struct {
	WriteURL string
	QueryURL string
	Client   *http.Client
}

// NewPrometheus builds a store for a base URL (e.g. http://tsdb:9090), deriving
// the remote-write and query endpoints from it.
func NewPrometheus(base string) *Prometheus {
	base = strings.TrimRight(base, "/")
	return &Prometheus{
		WriteURL: base + "/api/v1/write",
		QueryURL: base + "/api/v1/query",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Write sends the current fleet as a snappy-framed remote-write request.
func (p *Prometheus) Write(ctx context.Context, views []domain.HostView, now time.Time) error {
	ts := now.UnixMilli()
	req := &promv1.WriteRequest{}
	for _, s := range observe.SeriesOf(views) {
		labels := []*promv1.Label{{Name: "__name__", Value: s.Name}}
		for _, k := range sortedKeys(s.Labels) {
			labels = append(labels, &promv1.Label{Name: k, Value: s.Labels[k]})
		}
		req.Timeseries = append(req.Timeseries, &promv1.TimeSeries{
			Labels:  labels,
			Samples: []*promv1.Sample{{Value: s.Value, Timestamp: ts}},
		})
	}
	if len(req.Timeseries) == 0 {
		return nil
	}
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.WriteURL, bytes.NewReader(snappy.Encode(nil, body)))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	resp, err := p.client().Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("remote-write %s: %s", p.WriteURL, resp.Status)
	}
	return nil
}

// Restore reconstructs the last-known fleet from the latest heimdall_* samples.
func (p *Prometheus) Restore(ctx context.Context) ([]domain.HostView, error) {
	u := p.QueryURL + "?query=" + url.QueryEscape(`{__name__=~"heimdall_.+"}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("query %s: %s: %s", p.QueryURL, resp.Status, b)
	}
	var qr queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return nil, err
	}
	return reconstruct(qr.Data.Result), nil
}

func (p *Prometheus) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

type queryResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []vectorSample `json:"result"`
	} `json:"data"`
}

// vectorSample is one PromQL instant-vector element: a label set and a
// [unixSeconds, "value"] pair.
type vectorSample struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

// reconstruct groups samples by host and rebuilds best-effort host views: scalar
// gauges become OK metrics, series labels become tags, and the newest sample
// timestamp becomes last-seen. host_up provides existence (so offline hosts are
// kept); per-core and info series are skipped.
func reconstruct(results []vectorSample) []domain.HostView {
	type acc struct {
		labels   map[string]string
		metrics  []domain.Metric
		lastSeen time.Time
	}
	hosts := map[string]*acc{}
	for _, r := range results {
		id := r.Metric["host"]
		if id == "" {
			continue
		}
		a := hosts[id]
		if a == nil {
			a = &acc{labels: map[string]string{}}
			hosts[id] = a
		}
		if ts, ok := sampleTime(r.Value); ok && ts.After(a.lastSeen) {
			a.lastSeen = ts
		}
		for k, v := range r.Metric {
			switch k {
			case "__name__", "host", "state", "core":
			default:
				a.labels[k] = v
			}
		}
		name := r.Metric["__name__"]
		if name == "heimdall_host_up" {
			continue
		}
		if _, perCore := r.Metric["core"]; perCore {
			continue
		}
		if val, ok := sampleValue(r.Value); ok {
			a.metrics = append(a.metrics, domain.Metric{Name: restoreName(name), Status: domain.StatusOK, Gauge: val})
		}
	}
	out := make([]domain.HostView, 0, len(hosts))
	for id, a := range hosts {
		out = append(out, domain.HostView{
			Host: domain.Host{ID: domain.HostID(id), Hostname: id, DisplayName: id,
				Context: domain.HostContext{Labels: a.labels}},
			LastSeen:     a.lastSeen,
			LastSnapshot: a.metrics,
		})
	}
	return out
}

// restoreName maps a Prometheus series name back to a Heimdall metric name
// (heimdall_cpu_util -> cpu.util).
func restoreName(prom string) string {
	return strings.ReplaceAll(strings.TrimPrefix(prom, "heimdall_"), "_", ".")
}

func sampleValue(v []any) (float64, bool) {
	if len(v) != 2 {
		return 0, false
	}
	s, ok := v[1].(string)
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

func sampleTime(v []any) (time.Time, bool) {
	if len(v) != 2 {
		return time.Time{}, false
	}
	secs, ok := v[0].(float64)
	if !ok {
		return time.Time{}, false
	}
	return time.UnixMilli(int64(secs * 1000)), true
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
