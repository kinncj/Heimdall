// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package alert

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWebhookPostsEventJSON(t *testing.T) {
	got := make(chan map[string]any, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		got <- m
	}))
	defer srv.Close()

	Webhook{URL: srv.URL}.Notify(context.Background(), Event{
		Rule: "hot-cpu", Host: "a", State: Firing, Metric: "cpu.util", Value: 95, At: time.Unix(1_000, 0),
	})

	select {
	case m := <-got:
		if m["rule"] != "hot-cpu" || m["host"] != "a" || m["state"] != "firing" {
			t.Fatalf("webhook payload wrong: %v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was not called")
	}
}

func TestLoadRulesParsesDuration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.json")
	_ = os.WriteFile(path, []byte(`[
	  {"name":"hot","metric":"cpu.util","op":">","threshold":90,"for":"5m","match":{"env":"prod"}},
	  {"name":"low-disk","metric":"disk.used","op":">=","threshold":95}
	]`), 0o644)

	rules, err := LoadRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Fatalf("got %d rules, want 2", len(rules))
	}
	if rules[0].For != 5*time.Minute || rules[0].Match["env"] != "prod" {
		t.Errorf("rule 0 = %+v", rules[0])
	}
	if rules[1].For != 0 || rules[1].Op != OpGE {
		t.Errorf("rule 1 = %+v", rules[1])
	}
}
