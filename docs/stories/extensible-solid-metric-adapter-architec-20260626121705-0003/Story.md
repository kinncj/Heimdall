---
id: "extensible-solid-metric-adapter-architec-0003"
title: "Extensible SOLID Metric Adapter Architecture"
epic: "extensible-solid-metric-adapter-architec"
priority: "high"
ui: true
adr_required: true
milestone: null
labels:
  - "type:feature"
  - "priority:high"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-26T12:17:05-04:00"
---

# Extensible SOLID Metric Adapter Architecture

## Scenarios

```gherkin
Feature: Adapter-based metric collection design

  @story @priority:high
  Scenario: New metric is added through an adapter
    Given the monitoring system uses a metric adapter interface in both daemon and central processing components
    When a developer adds a new metric adapter for a new signal
    Then the new metric can be collected and displayed without changing existing adapters
    And the design follows SOLID principles with clear responsibilities and replaceable implementations

  Scenario: Adapter failure is isolated
    Given multiple metric adapters are active for a host
    When one adapter fails to collect a metric
    Then the failure is reported for that metric only
    And other adapters continue collecting and sending their metrics normally

  Scenario: Adapter reports unavailable vs insufficient-permission distinctly
    Given a metric adapter cannot collect its signal because the metric is unsupported on the host
    And another metric adapter cannot collect its signal because it lacks the required permission
    When the adapters report their status to the dashboard
    Then the unsupported metric is shown as unavailable with an em dash placeholder rather than an error
    And the permission-limited metric is shown as needing the privileged helper rather than an error
    And neither status is presented as a collection failure

  Scenario: Same adapter contract is shared by daemon and hub via the common versioned schema
    Given the daemon and the hub both depend on the shared metric adapter contract
    And the contract is defined by a common versioned schema
    When a metric adapter is built into both the daemon and the hub
    Then both components use the identical adapter interface and metric definitions
    And a schema version change is applied in one place and consumed by both components
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
