Feature: In-dashboard process view and privileged commands (v2 socket transport)

  A daemon can push a process table to the hub for the dashboard's top view, and
  expose on-demand commands. Privileged commands are delegated to the root helper;
  with no helper, the host says so rather than running them unprivileged.

  Scenario: A daemon pushes a process table the CLI can read
    Given a hub and a daemon pushing a process table
    When the operator reads the host's process table
    Then the process table lists running processes

  Scenario: A privileged command needs the helper when none is running
    Given a hub and a command-enabled daemon without a helper
    When the operator runs a privileged command
    Then the host reports that the command needs the privileged helper
