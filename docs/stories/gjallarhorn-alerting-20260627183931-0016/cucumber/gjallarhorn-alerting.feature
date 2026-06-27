Feature: Threshold alerting and notifications

  @story @priority:high
  Scenario: A breached rule fires an alert shown on the dashboard
    Given an alert rule of cpu.util above 90% for 5m
    When a host breaches 90% CPU continuously for 5 minutes
    Then the rule fires an alert
    And the dashboard shows the alert

  @story @priority:high
  Scenario: A webhook is posted on fire and again on clear
    Given an alert rule with a webhook configured
    When the alert fires and later clears
    Then the hub posts a notification to the webhook when it fires
    And the hub posts a notification to the webhook again when it clears

  @story @priority:high
  Scenario: A brief spike under the for-duration does not fire
    Given an alert rule of cpu.util above 90% for 5m
    When a host spikes above 90% briefly and recovers within the for-duration
    Then no alert fires

  @story @priority:high
  Scenario: Alerts can be scoped by tag
    Given an alert rule scoped to the tag env=prod
    When a host tagged env=prod breaches the threshold
    Then the alert fires for that host
    And hosts without the tag are not evaluated by the rule
