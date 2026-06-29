---
id: "dashboard-host-observability-modals-0024"
# ui:false — additive modals inside the existing detail view, built from existing
# TUI primitives (list, scroll, table). No new visual identity. The interaction
# design is documented in ADR 0017 and guides 06/07 (the "document everything" ask)
# rather than the wireframe/mockup design gate. Gated by tests.
title: "Seamless in-dashboard host observability — log + top modals (Heimdallr's sight)"
epic: "heimdallr-sight"
priority: "high"
ui: false
adr_required: true
milestone: "v1.6.0"
phase: done
labels:
  - "type:feature"
  - "priority:high"
status: done
issue_number: null
issue_url: null
created_at: "2026-06-29T04:03:44+0000"
---

# Seamless in-dashboard host observability — log + top modals

## Story

**As an** operator watching a fleet,
**I want** the dashboard to discover which hosts expose logs and a control plane
and let me read their logs and live process table from the host detail view,
**so that** I triage a host without leaving the dashboard or typing addresses,
tokens, or `--run` flags.

## Context

Today the control plane (`ControlPlaneService`) and log streaming
(`LogStreamService`) live on the daemon's `--control-listen` port and are reached
only by the dashboard connecting **directly to the daemon**
(`heimdall-dashboard --control <addr> --run <cmd>`). That breaks the fleet model:
only hubs listen; daemons and dashboards are outbound-only and need no inbound
ports (guide 05, ADR 0006), so a daemon is usually not reachable from a dashboard.

This story makes logs/control **hub-mediated**, like metrics: the daemon serves
them in reverse over an outbound channel to the hub, and the dashboard requests
them **from the hub**. See ADR 0017.

## Acceptance Criteria

```gherkin
@story:dashboard-host-observability-modals-0024 @epic:heimdallr-sight @priority:high
Feature: Seamless in-dashboard host observability

  Scenario: A daemon advertises its log sources and process table
    Given a daemon configured with log sources and process push
    When it pushes to the hub
    Then the hub buffers the host's log lines and latest process table and marks
      the host as exposing those capabilities

  Scenario: Daemons that push nothing are unaffected
    Given a daemon configured with no log sources and no process push
    When it streams to the hub
    Then the host exposes no observability capability and the dashboard shows no
      log or top affordance for it

  Scenario: The detail view advertises available observability
    Given the dashboard detail view of a host that exposes logs and control
    Then the footer offers the log and top keybinds

  Scenario: Opening logs lists the available sources
    Given the detail view of a host exposing two log sources
    When the operator presses the log key
    Then a modal lists the two log sources for selection

  Scenario: Selecting a source streams it, scrollable
    Given the log-source list is open
    When the operator selects a source with enter
    Then the same modal streams that source's lines and scrolls

  Scenario: Escape is the back button at every level
    Given a log source is streaming in the modal
    When the operator presses escape
    Then the modal returns to the log-source list
    And pressing escape again returns to the host detail view

  Scenario: Top shows the live process table
    Given the detail view of a host that exposes a control plane
    When the operator presses the top key
    Then a modal shows the host's process table and refreshes while open
    And the table scrolls

  Scenario: The process table works across operating systems
    Given hosts running Linux, macOS, and Windows
    When the operator opens top for each
    Then each shows that OS's process listing without error
```

## Definition of Done

- [ ] Unit tests green (capability advertisement + hub propagation + modal state machine + cross-OS process command)
- [ ] `make test-all` green
- [ ] Backward compatible: additive labels; old peers ignore them; no control plane ⇒ no affordance
- [ ] SOLID: capability carried as data; commands/columns as registries (no switches)
- [ ] ADR 0017 written; guides 06 (control plane) and 07 (log streaming) updated
- [ ] CHANGELOG entry under [1.6.0]
- [ ] Released as v1.6.0

## The one accepted breaking change

The daemon stops acting as a server. Removed: daemon `--control-listen`,
`--control-token`, `--control-tls-cert`, `--control-tls-key`; dashboard
`--control`, `--run`, `--tail`. `--log-source` is kept but now configures what the
daemon **pushes**. This is the single authorized backward-compatibility break.

## Non-goals

- Writing/mutating commands — the process command stays read-only and allow-listed.
- On-demand / interactive execution against a daemon — there is no request path to
  the daemon in v1 (that is the v2 socket model in ADR 0017 §3.9).
- A new identity system — the hub's existing auth covers daemons and dashboards.
- Historical log search; the modal shows the hub's live + buffered tail.

## Design

See ADR 0017. Key points:
- The daemon is a pure outbound producer: it pushes tailed log lines and a periodic
  process table to the hub on its existing stream. It never listens.
- The hub buffers a bounded log ring + the latest process table per host and serves
  them to dashboards; capability is advertised as additive reserved labels
  (`_logs`, hub-stamped `_control`) — never an address.
- The dashboard talks only to the hub; the `l`/`t` modals are sub-models reached
  from the detail view, with escape unwinding one level at a time.
