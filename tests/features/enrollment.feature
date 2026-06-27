Feature: Remote hardware daemon enrollment and streaming

  Scenario: Daemon connects from mixed operating systems
    Given a lightweight daemon is installed on Windows, macOS, and Linux hosts
    And the hosts include devices such as a workstation, DGX Spark, HP Strix Halo, Mac mini, Raspberry Pi, and Alienware machine
    When each daemon starts and connects to the central service over a low-bandwidth socket protocol
    Then each host appears as an online monitor target in the centralized system
    And each host sends periodic metric updates without requiring high network throughput

  Scenario: Enrollment over TLS requires a valid enrollment token
    Given the central hub requires TLS and a valid enrollment token to register a daemon
    When a daemon attempts to enroll over TLS using an invalid or missing enrollment token
    Then the hub rejects the connection during enrollment
    And the unauthenticated daemon is not registered as a monitor target
