Feature: Hub-mediated allow-listed commands (v2 socket transport)

  On-demand, read-only, allow-listed commands run through the hub: the dashboard or
  CLI asks the hub, which routes the request down the daemon's outbound stream; the
  daemon runs it as its unprivileged user and returns the result. No daemon listens.

  Scenario: Operator runs an allow-listed read-only command
    Given a hub and a command-enabled daemon
    When the operator runs an allow-listed command via the CLI
    Then the command runs on the host and the JSON result is returned

  Scenario: Non-allow-listed commands are refused
    Given a hub and a command-enabled daemon
    When the operator runs a command that is not on the allow-list
    Then the host refuses it with insufficient_permission and runs nothing

  Scenario: A daemon without --allow-commands exposes no command surface
    Given a hub and a daemon with commands disabled
    When the operator runs an allow-listed command via the CLI
    Then the host refuses it because commands are disabled

  Scenario: Every command is audit-logged on the daemon
    Given a hub and a command-enabled daemon
    When the operator runs an allow-listed command via the CLI
    Then the daemon records an audit entry naming the command and the operator
