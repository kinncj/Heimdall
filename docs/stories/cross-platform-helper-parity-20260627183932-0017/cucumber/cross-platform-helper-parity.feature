Feature: Cross-platform privileged metrics parity

  @story @priority:medium
  Scenario: Linux reports CPU package power and temperatures
    Given a privileged helper runs on a Linux host
    When the helper reports power and thermal metrics
    Then the helper reports CPU package power via RAPL
    And the helper reports temperatures via hwmon

  @story @priority:medium
  Scenario: An unsupported metric reports unavailable instead of failing
    Given a Linux host without RAPL support
    When the helper reads CPU package power
    Then the metric reports unavailable or needs-helper
    And the daemon keeps reporting its other metrics

  @story @priority:medium
  Scenario: Windows has a privileged path for power and thermal metrics
    Given a privileged helper runs on a Windows host
    When the helper reports power and thermal metrics
    Then a privileged path provides power and thermal metrics comparable to Apple Silicon
