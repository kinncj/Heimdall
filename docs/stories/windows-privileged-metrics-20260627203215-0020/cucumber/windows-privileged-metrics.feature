Feature: Windows privileged power and thermal metrics

  @story @priority:medium
  Scenario: Report CPU zone temperature on Windows via WMI
    Given the helper runs on a Windows host with WMI available
    When the helper collects privileged metrics
    Then it reports the CPU zone temperature as temp.pkg

  @story @priority:medium
  Scenario: CPU package power is unavailable on Windows
    Given the helper runs on a Windows host
    When the helper collects privileged metrics
    Then it reports CPU package power as unavailable because RAPL is absent

  @story @priority:medium
  Scenario: Metrics degrade to unavailable when WMI is absent
    Given the helper runs on a Windows host where WMI is unavailable
    When the helper collects privileged metrics
    Then each affected metric degrades to unavailable instead of failing
