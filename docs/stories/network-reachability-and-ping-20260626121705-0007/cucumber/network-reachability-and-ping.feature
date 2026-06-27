Feature: Host network reachability and latency monitoring

  @story @priority:medium
  Scenario: Daemon measures latency and internet reachability
    Given a daemon is configured with a reachability target
    When the daemon probes the configured target and the public internet
    Then it reports the measured latency to the configured target
    And it reports internet reachability as up or down

  Scenario: Dashboard shows reachability state and latency trend
    Given a host is reporting reachability and latency metrics
    When the operator views that host in the dashboard
    Then the dashboard shows the reachability state with both a symbol and text
    And the dashboard shows a latency trend over recent history

  Scenario: A failed reachability probe is isolated
    Given a host is online and streaming metrics
    When the reachability probe fails for that host
    Then the host remains shown as online
    And only the reachability metric is shown in an error state
    And the other metrics for that host keep updating normally
