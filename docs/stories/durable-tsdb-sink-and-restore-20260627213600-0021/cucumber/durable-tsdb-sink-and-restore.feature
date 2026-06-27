Feature: Durable TSDB sink and fleet restore

  @story @priority:high
  Scenario: Hub writes metrics to a configured TSDB continuously
    Given the hub is started with a Prometheus-compatible TSDB configured
    When daemons report metrics to the hub
    Then the hub writes those metrics to the configured TSDB continuously

  @story @priority:high
  Scenario: Hub restores the last-known fleet from the TSDB after a restart
    Given the hub previously persisted a fleet to the configured TSDB
    When the hub restarts before any daemon reconnects
    Then the hub restores the last-known fleet from the configured TSDB
    And offline hosts are restored with their last-seen age

  @story @priority:high
  Scenario: Hub stays in-memory and loses state without a configured TSDB
    Given the hub is started with no TSDB configured
    When the hub restarts
    Then the hub starts with an empty fleet and no restored state

  @story @priority:high
  Scenario: Restore is best-effort and excludes info strings and alert state
    Given the hub restores a fleet from the configured TSDB
    When the restored fleet is presented before live data resumes
    Then info strings and alert state are absent from the restored fleet
    And they reappear once live data resumes
