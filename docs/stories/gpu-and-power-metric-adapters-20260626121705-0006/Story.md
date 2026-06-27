---
id: "gpu-and-power-metric-adapters-0006"
title: "GPU and Power Metric Adapters"
epic: "gpu-and-power-metric-adapters"
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

# GPU and Power Metric Adapters

## Scenarios

```gherkin
Feature: GPU and power metric adapters across vendors

  @story @priority:high
  Scenario: NVIDIA GPU metrics are collected through NVML
    Given a host has an NVIDIA GPU and the NVML library is available
    When the GPU adapter collects metrics through NVML
    Then it reports GPU utilization, VRAM usage, temperature, and power draw
    And these values appear for that host in the dashboard

  Scenario: Apple Silicon GPU and power metrics are collected through the platform path
    Given a host is an Apple Silicon Mac with the privileged helper installed
    When the GPU adapter collects metrics through the Apple Silicon platform path
    Then it reports GPU and power metrics for that host
    And the daemon collects them without requiring elevated privileges itself

  Scenario: Unsupported GPU vendor degrades gracefully
    Given a host has a GPU from an unsupported vendor such as a Raspberry Pi or an unsupported AMD card
    When the GPU adapter attempts to collect metrics
    Then the adapter reports the GPU metrics as unavailable
    And the daemon continues collecting other metrics without crashing

  Scenario: Power profile is displayed read-only
    Given a host reports a power profile through the power adapter
    When the dashboard displays the host power information
    Then the power profile is shown as read-only information
    And the dashboard offers no control to change the power profile
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
