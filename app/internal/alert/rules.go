// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package alert

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ruleFile is the on-disk shape of a rule: For is a human duration string
// ("5m", "30s") rather than nanoseconds.
type ruleFile struct {
	Name      string            `json:"name"`
	Metric    string            `json:"metric"`
	Op        Op                `json:"op"`
	Threshold float64           `json:"threshold"`
	For       string            `json:"for"`
	Match     map[string]string `json:"match,omitempty"`
}

// LoadRules reads a JSON array of rules from path. For is parsed as a duration
// string; an empty For means fire immediately on breach.
func LoadRules(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []ruleFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("alert: parse rules: %w", err)
	}
	rules := make([]Rule, 0, len(raw))
	for i, rf := range raw {
		var dur time.Duration
		if rf.For != "" {
			dur, err = time.ParseDuration(rf.For)
			if err != nil {
				return nil, fmt.Errorf("alert: rule %d (%q) bad for=%q: %w", i, rf.Name, rf.For, err)
			}
		}
		rules = append(rules, Rule{
			Name: rf.Name, Metric: rf.Metric, Op: rf.Op,
			Threshold: rf.Threshold, For: dur, Match: rf.Match,
		})
	}
	return rules, nil
}
