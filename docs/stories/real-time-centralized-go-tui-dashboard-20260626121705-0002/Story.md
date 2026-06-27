---
id: "real-time-centralized-go-tui-dashboard-0002"
title: "Real-Time Centralized Go TUI Dashboard"
epic: "real-time-centralized-go-tui-dashboard"
priority: "high"
ui: true
adr_required: false
milestone: null
labels:
  - "type:feature"
  - "priority:high"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-26T12:17:05-04:00"
---

# Real-Time Centralized Go TUI Dashboard

## Scenarios

```gherkin
Feature: Centralized terminal dashboard for remote hardware monitoring

  @story @priority:high
  Scenario: Operator sees live metrics across many machines
    Given a Go-based TUI dashboard is running on a central machine
    And multiple remote daemons are streaming metrics
    When the operator opens the dashboard
    Then the dashboard shows each remote host in real time
    And the dashboard displays useful system data such as CPU, memory, storage, and temperature trends per host

  Scenario: Dashboard handles no-data and stale-data states
    Given one or more remote hosts stop sending updates
    When the data for those hosts becomes stale
    Then the dashboard clearly marks those hosts as stale or offline
    And the last known values remain visible with a clear timestamp to avoid misleading real-time status

  Scenario: Dashboard subscribes to the hub metric bus and renders all core metrics per host
    Given the dashboard is subscribed to the hub metric bus
    And hosts are publishing their metric streams to the hub
    When metric updates arrive for each host
    Then the dashboard renders CPU, memory, storage, temperature, GPU, power, and network for that host
    And each metric updates in place as new values arrive

  Scenario: Operator opens a host detail view with trend graphs
    Given the dashboard is showing several hosts in an overview
    When the operator selects a single host to inspect
    Then the dashboard opens a host detail view for that host
    And the detail view shows trend graphs built from in-memory metric history
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
