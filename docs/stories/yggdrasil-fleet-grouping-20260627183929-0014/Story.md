---
id: "yggdrasil-fleet-grouping-0014"
title: "Yggdrasil: Topology-Aware Fleet Grouping"
epic: "yggdrasil-fleet-grouping"
priority: "medium"
ui: true
adr_required: true
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

# Yggdrasil: Topology-Aware Fleet Grouping

## Story

As an operator,
I want to group, filter, and search the fleet by dimension,
so that large, federated fleets stay legible.

## Scenarios

```gherkin
Feature: Topology-aware fleet grouping

  @story @priority:medium
  Scenario: Group the grid by Bifrost origin hub
    Given the dashboard shows a federated fleet
    When the operator groups the grid by Bifrost origin hub
    Then each host appears under its origin edge hub

  @story @priority:medium
  Scenario: Filter the fleet by a tag
    Given the dashboard shows a fleet with tagged hosts
    When the operator filters by the tag env=prod
    Then only hosts tagged env=prod remain visible

  @story @priority:medium
  Scenario: Search by a host name that matches nothing shows an empty state
    Given the dashboard shows many hosts
    When the operator searches for a host name that matches no host
    Then the grid shows an empty state instead of any hosts

  @story @priority:medium
  Scenario: Group the grid by operating system
    Given the dashboard shows hosts running different operating systems
    When the operator groups the grid by OS
    Then hosts appear grouped under their operating system
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

ADR 0010
