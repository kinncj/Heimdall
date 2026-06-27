---
id: "optional-privileged-metrics-helper-0005"
title: "Optional Privileged Metrics Helper"
epic: "optional-privileged-metrics-helper"
priority: "medium"
ui: true
adr_required: true
milestone: null
labels:
  - "type:feature"
  - "priority:medium"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-26T12:17:05-04:00"
---

# Optional Privileged Metrics Helper

## Scenarios

```gherkin
Feature: Optional privileged helper for power and thermal metrics

  @story @priority:medium
  Scenario: Privileged metrics become available when the helper is installed
    Given the optional privileged helper is installed on a host
    And the metrics daemon is running as an unprivileged user
    When the daemon requests power and full thermal metrics through the helper
    Then the power and full thermal metrics are collected and reported
    And the daemon continues to run without elevated privileges

  Scenario: Privileged metrics report needing the helper when it is absent
    Given the optional privileged helper is not installed on a host
    When the power and full thermal adapters attempt to collect their metrics
    Then those adapters report insufficient permission rather than an error
    And the dashboard shows a needs-helper affordance for those metrics
    And the daemon keeps running without crashing

  Scenario: Helper runs as a separate privileged unit over a local socket
    Given the privileged helper runs as its own privileged unit on the host
    And the helper exposes privileged metrics to the daemon over a local socket
    When the unprivileged daemon collects privileged metrics
    Then the daemon obtains the values through the local socket
    And the daemon never invokes sudo or runs as root
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
