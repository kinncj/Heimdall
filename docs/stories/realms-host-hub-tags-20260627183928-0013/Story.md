---
id: "realms-host-hub-tags-0013"
title: "Realms: Host and Hub Tags"
epic: "realms-host-hub-tags"
priority: "medium"
ui: false
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

# Realms: Host and Hub Tags

## Story

As an operator,
I want to tag hosts and hubs,
so that I can organize and filter the fleet.

## Scenarios

```gherkin
Feature: Host and hub tags

  @story @priority:medium
  Scenario: A daemon's tags are shown in the dashboard
    Given a daemon is started with the tags env=prod and role=db
    When the daemon's metrics reach the dashboard
    Then the dashboard shows the host carrying the tags env=prod and role=db

  @story @priority:medium
  Scenario: A hub's tags inherit to the hosts it relays via Bifrost
    Given a hub carries the tag region=eu
    Given the hub relays several hosts over Bifrost
    When those relayed hosts appear in the dashboard
    Then each relayed host inherits the tag region=eu from its hub

  @story @priority:medium
  Scenario: A host's own tag overrides an inherited hub tag on conflict
    Given a hub carries the tag env=staging
    Given a host the hub relays is started with its own tag env=prod
    When the host's tags are resolved
    Then the host's own tag env=prod overrides the inherited hub tag
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

ADR 0010
