Feature: Opt-in host log streaming to the dashboard

  @story @priority:low
  Scenario: Daemon tails a configured log source and streams it
    Given a log source is configured on a host
    When the daemon tails the configured log source
    Then the daemon streams new log lines to the hub on a separate log stream
    And the log stream is kept independent of the metric stream

  Scenario: Dashboard shows live, rate-limited logs per host
    Given a host is streaming logs to the hub
    When the operator opens the logs pane for that host
    Then the logs pane shows live log lines for that host
    And the log stream is rate-limited so it does not overwhelm the low-bandwidth link

  Scenario: No logs are streamed when no log source is configured
    Given a host has no log source configured
    When the daemon runs and streams metrics
    Then no log lines are streamed for that host
    And log streaming stays off until a log source is explicitly configured
