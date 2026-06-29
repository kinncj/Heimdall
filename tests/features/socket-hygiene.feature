@story:unprivileged-terminal-control-plane-0010 @epic:heimdallr-sight @priority:high
Feature: Socket hygiene (v2 — daemons never listen; directives ride one stream)

  The v2 control plane reuses the daemon's single outbound stream to the hub for
  every directive (logs, top, commands). The security model demands that a daemon
  expose no inbound surface and that an on-demand directive open no new socket —
  it must ride the stream the daemon already holds. These scenarios prove it by
  inspecting the running processes' real sockets, not the code.

  Scenario: A daemon exposes no inbound socket
    Given a hub and a daemon pushing a process table
    Then the daemon process is listening on no network port
    And the daemon process is listening on no unix socket
    And the hub is the only one of our processes holding a listening network port

  Scenario: The daemon holds a single outbound connection, to the hub
    Given a hub and a daemon pushing a process table
    Then the daemon has exactly one established network connection, to the hub

  Scenario: An on-demand command opens no new socket — it rides the stream
    Given a hub and a command-enabled daemon without a helper
    And the daemon's established connections are recorded
    When the operator runs an allow-listed command on the host
    Then the command result returns
    And the daemon opened no new connection
