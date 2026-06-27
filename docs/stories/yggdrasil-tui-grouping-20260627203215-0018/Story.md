---
id: "yggdrasil-tui-grouping-0018"
title: "Yggdrasil: In-Dashboard Fleet Grouping, Filtering, and Search"
epic: "yggdrasil-tui-grouping"
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

# Yggdrasil: In-Dashboard Fleet Grouping, Filtering, and Search

## Story

As an operator,
I want to group, filter, and search the fleet from inside the dashboard,
so that a large, federated fleet stays legible.

## Scenarios

```gherkin
Feature: In-dashboard fleet grouping, filtering, and search

  @story @priority:high
  Scenario: Group the fleet by origin hub
    Given the dashboard shows a federated fleet
    When the operator presses the group key to group by origin hub
    Then hosts appear under per-hub section headers

  @story @priority:high
  Scenario: Cycle the group dimension to OS and then to a tag key
    Given the dashboard is grouped by origin hub
    When the operator cycles the group dimension to OS
    And the operator cycles the group dimension to a tag key
    Then hosts appear grouped under section headers for the active dimension

  @story @priority:high
  Scenario: Filter the fleet by host name or tag value
    Given the dashboard shows many hosts
    When the operator enters a filter term with the slash key
    Then only hosts matching the host name or a tag value remain visible

  @story @priority:high
  Scenario: A filter that matches nothing shows an empty state
    Given the dashboard shows many hosts
    When the operator enters a filter term that matches no host
    Then the dashboard shows an empty state instead of any hosts

  @story @priority:high
  Scenario: Grouping and filtering work in demo mode
    Given the dashboard runs in demo mode
    When the operator groups and filters the fleet
    Then grouping and filtering behave the same as with a live fleet
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

ADR 0013
