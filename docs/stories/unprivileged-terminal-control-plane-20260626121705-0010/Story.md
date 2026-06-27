---
id: "unprivileged-terminal-control-plane-0010"
title: "Unprivileged Terminal Control Plane"
epic: "unprivileged-terminal-control-plane"
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

# Unprivileged Terminal Control Plane

## Scenarios

```gherkin
Feature: Read-only allow-listed remote control plane

  @story @priority:medium
  Scenario: Operator runs an allow-listed read-only query
    Given the control plane exposes an allow-list of read-only commands on a host
    When the operator runs an allow-listed query such as listing processes, showing disk usage, or listing files in an allowed directory
    Then the host runs the query as the unprivileged user and returns the result
    And the result is shown in the dashboard

  Scenario: Escalation and non-allow-listed commands are refused
    Given the control plane runs commands as the unprivileged daemon user
    When the operator attempts to use sudo or run a command that is not on the allow-list
    Then the host refuses the command
    And no command is run with elevated privileges

  Scenario: Every control-plane invocation is audit-logged
    Given the control plane is enabled on a host
    When any control-plane command is invoked
    Then the host records an audit log entry for that invocation
    And the audit entry identifies the command and the requesting operator
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
