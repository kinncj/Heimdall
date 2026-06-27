---
id: "windows-privileged-metrics-0020"
title: "Windows Privileged Power and Thermal Metrics"
epic: "windows-privileged-metrics"
priority: "medium"
ui: false
adr_required: true
milestone: "v1.3.0"
phase: discover
labels:
  - "type:feature"
  - "priority:medium"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-06-27T00:00:00+0000"
---

# Windows Privileged Power and Thermal Metrics

## Story

As an operator,
I want power and thermal metrics on Windows like Linux and macOS,
so that fleet observability is consistent across platforms.

## Scenarios

```gherkin
Feature: Windows privileged power and thermal metrics

  @story @priority:medium
  Scenario: Report CPU zone temperature on Windows via WMI
    Given the helper runs on a Windows host with WMI available
    When the helper collects privileged metrics
    Then it reports the CPU zone temperature as temp.pkg

  @story @priority:medium
  Scenario: CPU package power is unavailable on Windows
    Given the helper runs on a Windows host
    When the helper collects privileged metrics
    Then it reports CPU package power as unavailable because RAPL is absent

  @story @priority:medium
  Scenario: Metrics degrade to unavailable when WMI is absent
    Given the helper runs on a Windows host where WMI is unavailable
    When the helper collects privileged metrics
    Then each affected metric degrades to unavailable instead of failing
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber/Behave scenarios green
- [ ] ADRs linked where required
- [ ] CHANGELOG entry added
- [ ] PR description references this story

## ADR Links

ADR 0015
