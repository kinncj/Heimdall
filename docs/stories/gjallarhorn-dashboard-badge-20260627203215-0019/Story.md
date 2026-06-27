---
id: "gjallarhorn-dashboard-badge-0019"
title: "Gjallarhorn: Firing Alerts Visible in the Dashboard"
epic: "gjallarhorn-dashboard-badge"
priority: "high"
ui: true
adr_required: true
milestone: "v1.3.0"
phase: discover
labels:
  - "type:feature"
  - "priority:high"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-27T00:00:00+0000"
---

# Gjallarhorn: Firing Alerts Visible in the Dashboard

## Story

As an operator,
I want firing alerts surfaced in the dashboard,
so that I notice them without watching the logs.

## Scenarios

```gherkin
Feature: Firing alerts visible in the dashboard

  @story @priority:high
  Scenario: A host with a firing alert is highlighted
    Given the dashboard shows a fleet
    And one host has a firing alert
    Then that host's row shows an alert badge or highlight

  @story @priority:high
  Scenario: The header shows the fleet alert count
    Given the dashboard shows a fleet with firing alerts on several hosts
    Then the header shows the count of hosts with firing alerts

  @story @priority:high
  Scenario: The badge clears when the alert resolves
    Given a host's row shows an alert badge
    When that host's alert clears
    Then the badge disappears from that host's row

  @story @priority:high
  Scenario: Alerts are visible in demo mode
    Given the dashboard runs in demo mode
    And a host has a firing alert
    Then the alert badge and the fleet alert count are visible
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required
- [ ] Wireframe approved
- [ ] Mockup approved
- [ ] Accessibility audit passed

## ADR Links

ADR 0014
