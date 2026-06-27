// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Notifier delivers an alert event to an external sink.
type Notifier interface {
	Notify(ctx context.Context, ev Event)
}

// Webhook posts each event as a JSON object to a URL (Slack/PagerDuty-compatible
// generic shape). Delivery is best-effort; failures are reported via OnError.
type Webhook struct {
	URL     string
	Client  *http.Client
	OnError func(error)
}

// Notify POSTs the event as JSON. It never blocks longer than the client timeout.
func (w Webhook) Notify(ctx context.Context, ev Event) {
	body, _ := json.Marshal(map[string]any{
		"rule":   ev.Rule,
		"host":   ev.Host,
		"state":  string(ev.State),
		"metric": ev.Metric,
		"value":  ev.Value,
		"at":     ev.At.UTC().Format(time.RFC3339),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		w.reportError(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := w.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		w.reportError(err)
		return
	}
	_ = resp.Body.Close()
}

func (w Webhook) reportError(err error) {
	if w.OnError != nil {
		w.OnError(err)
	}
}
