// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package observe

import (
	"encoding/json"
	"net/http"

	"heimdall/app/internal/domain"
)

// Source is the read side of the hub registry the exporter needs.
type Source interface {
	Hosts() []domain.HostView
}

// Handler builds the Mímir HTTP surface: GET /metrics (OpenMetrics text for a
// Prometheus scraper) and GET /history?host=&metric= (JSON trend samples). hist
// may be nil to disable the history endpoint.
func Handler(src Source, hist *History) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(RenderOpenMetrics(src.Hosts())))
	})
	mux.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		if hist == nil {
			http.Error(w, "history disabled", http.StatusNotFound)
			return
		}
		host, metric := r.URL.Query().Get("host"), r.URL.Query().Get("metric")
		if host == "" || metric == "" {
			http.Error(w, "host and metric query params required", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(hist.Series(host, metric))
	})
	return mux
}
