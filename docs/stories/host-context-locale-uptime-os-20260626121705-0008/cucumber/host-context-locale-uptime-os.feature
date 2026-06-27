Feature: Host context reporting and display

  @story @priority:low
  Scenario: Daemon reports host context on enrollment and periodically
    Given a daemon is enrolling with the hub
    When the daemon sends its host context
    Then the context includes operating system and architecture, hostname, locale and timezone, uptime, and boot time
    And the daemon refreshes the host context periodically while connected

  Scenario: Dashboard shows host context in host detail
    Given a host has reported its host context
    When the operator opens the host detail view
    Then the dashboard shows the operating system and architecture, hostname, locale and timezone, uptime, and boot time

  Scenario: Host context updates without re-registering the host
    Given a host is already enrolled and visible in the dashboard
    When the host context changes and the daemon sends an update
    Then the dashboard reflects the updated context for the same host
    And the host is not re-registered or duplicated
