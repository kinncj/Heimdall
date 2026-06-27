---
id: "cross-platform-helper-parity-0017"
title: "Cross-Platform Privileged Metrics Parity"
epic: "cross-platform-helper-parity"
priority: "medium"
ui: false
adr_required: false
milestone: "v1.2.0"
phase: discover
labels:
  - "type:feature"
  - "priority:medium"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-27T00:00:00+0000"
---

# Cross-Platform Privileged Metrics Parity

## Story

As an operator,
I want power and thermal metrics on Linux and Windows like on Apple Silicon,
so that coverage is consistent across the fleet.

## Scenarios

```gherkin
Feature: Cross-platform privileged metrics parity

  @story @priority:medium
  Scenario: Linux reports CPU package power and temperatures
    Given a privileged helper runs on a Linux host
    When the helper reports power and thermal metrics
    Then the helper reports CPU package power via RAPL
    And the helper reports temperatures via hwmon

  @story @priority:medium
  Scenario: An unsupported metric reports unavailable instead of failing
    Given a Linux host without RAPL support
    When the helper reads CPU package power
    Then the metric reports unavailable or needs-helper
    And the daemon keeps reporting its other metrics

  @story @priority:medium
  Scenario: Windows has a privileged path for power and thermal metrics
    Given a privileged helper runs on a Windows host
    When the helper reports power and thermal metrics
    Then a privileged path provides power and thermal metrics comparable to Apple Silicon
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

ADR 0004, ADR 0005
