---
id: "mimir-history-export-0015"
title: "Mimir: Metrics History and OpenMetrics Export"
epic: "mimir-history-export"
priority: "high"
ui: false
adr_required: true
milestone: "v1.2.0"
phase: discover
labels:
  - "type:feature"
  - "priority:high"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-27T00:00:00+0000"
---

# Mimir: Metrics History and OpenMetrics Export

## Story

As an SRE,
I want the hub to expose metrics for Prometheus and keep short-range history,
so that I can integrate with existing tooling and see recent trends.

## Scenarios

```gherkin
Feature: Metrics history and OpenMetrics export

  @story @priority:high
  Scenario: A Prometheus scraper reads the hub's OpenMetrics endpoint
    Given a hub is collecting fleet metrics
    When a Prometheus scraper reads the hub's /metrics endpoint
    Then the scraper receives the metrics in OpenMetrics format

  @story @priority:high
  Scenario: Exported series carry host, origin hub, and tag labels
    Given the hub exports metrics in OpenMetrics format
    When a series is exported for a host
    Then the series carries the host id, origin hub, and tags as labels

  @story @priority:high
  Scenario: Recent samples are served as short-range trends
    Given the hub retains recent samples in a bounded in-memory window
    When an operator requests recent trends for a host
    Then the hub serves the retained samples as a short-range trend

  @story @priority:high
  Scenario: History is lost on hub restart as documented
    Given the hub holds recent samples only in memory
    When the hub restarts
    Then the prior history is gone
    And this loss is documented as acceptable
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

ADR 0011
