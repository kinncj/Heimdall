# SPDX-License-Identifier: AGPL-3.0-or-later
# Story: docs/stories/top-view-core-topology-20260702184928-0028/story.md
Feature: Per-core topology and per-cluster CPU power
  The daemon labels each logical core with its type (P/E on hybrid silicon,
  uniform elsewhere) and, where the SMC exposes per-cluster rails (Apple
  Pro/Max), decomposes CPU power into P/E cluster watts. Both are new metric
  names only — an older dashboard simply ignores them.

  Scenario: The daemon emits a topology consistent with its per-core readings
    Given a built heimdall-daemon binary
    When the daemon collects one snapshot
    Then the snapshot contains an OK "cpu.topology" metric
    And the topology has exactly one entry per logical core
    And the topology detail describes the core mix
    And every topology entry is a known core-type id

  Scenario: Cluster power rails are consistent when present
    Given a built heimdall-daemon binary
    When the daemon collects one snapshot
    Then any "power.cpu.pcluster" and "power.cpu.ecluster" readings are non-negative watts
    And when both cluster rails are OK their sum is close to "power.cpu"
