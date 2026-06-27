---
id: "centralized-dashboard-federation-relay-0009"
title: "Centralized Dashboard Federation Relay"
epic: "centralized-dashboard-federation-relay"
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

# Centralized Dashboard Federation Relay

## Scenarios

```gherkin
Feature: Hub federation and metric relay

  @story @priority:medium
  Scenario: A hub relays its metric stream to a parent hub
    Given a local hub is collecting metrics from its hosts
    And a parent cloud hub is configured as an upstream
    When the local hub connects to the parent hub as a subscriber
    Then the parent hub receives the relayed metric stream for the local hosts
    And operators on the parent hub can monitor those hosts

  Scenario: Multiple dashboards subscribe to one hub concurrently
    Given a hub is publishing host metrics
    When several dashboards subscribe to the same hub at the same time
    Then every subscribed dashboard receives the host metrics
    And all dashboards show consistent data for the same hosts

  Scenario: Cross-hub link is authenticated and resumes cleanly after a drop
    Given a local hub is relaying metrics to an authenticated parent hub
    When the cross-hub link drops and later reconnects
    Then the parent hub re-authenticates the local hub before accepting data
    And host data resumes without duplication or corruption
    And the relay does not create a loop between hubs
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
