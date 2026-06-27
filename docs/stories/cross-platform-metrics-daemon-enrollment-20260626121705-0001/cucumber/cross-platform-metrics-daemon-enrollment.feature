Feature: Remote hardware daemon enrollment and streaming

  @story @priority:high
  Scenario: Daemon connects from mixed operating systems
    Given a lightweight daemon is installed on Windows, macOS, and Linux hosts
    And the hosts include devices such as a workstation, DGX Spark, HP Strix Halo, Mac mini, Raspberry Pi, and Alienware machine
    When each daemon starts and connects to the central service over a low-bandwidth socket protocol
    Then each host appears as an online monitor target in the centralized system
    And each host sends periodic metric updates without requiring high network throughput

  Scenario: Daemon loses connection and reconnects
    Given a registered host daemon was previously streaming metrics
    When the network connection drops and later returns
    Then the host is marked offline during the outage
    And the daemon reconnects automatically and resumes metric streaming without duplicate host registration

  Scenario: Enrollment over TLS requires a valid enrollment token
    Given the central hub requires TLS and a valid enrollment token to register a daemon
    When a daemon attempts to enroll over TLS using an invalid or missing enrollment token
    Then the hub rejects the connection during enrollment
    And the unauthenticated daemon is not registered as a monitor target

  Scenario: Reconnect resumes with the same stable HostID without duplicate registration
    Given an enrolled daemon has a stable HostID assigned at first enrollment
    And the daemon has been streaming metrics under that HostID
    When the daemon restarts and reconnects to the hub
    Then the daemon re-presents the same stable HostID
    And the hub matches it to the existing host instead of creating a new registration
    And no duplicate host entry appears in the dashboard

  Scenario: Daemon streams at a configurable low-bandwidth interval
    Given a daemon is configured with a metric streaming interval
    When the operator sets a longer interval to conserve bandwidth
    Then the daemon emits metric updates at the configured interval
    And the stream stays within a low-bandwidth budget suitable for constrained links
