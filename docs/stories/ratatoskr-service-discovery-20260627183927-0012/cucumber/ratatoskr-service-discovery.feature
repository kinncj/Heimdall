Feature: Daemon auto-discovery of its hub

  @story @priority:high
  Scenario: A daemon with discovery enabled finds and streams to an advertised hub
    Given a hub is advertising itself on the local network
    Given a daemon has discovery enabled and no hub is configured
    When the daemon starts
    Then the daemon discovers the advertised hub
    And the daemon streams its metrics to that hub without a hub being hand-configured

  @story @priority:high
  Scenario: An explicit hub always wins over discovery
    Given a hub is advertising itself on the local network
    Given a daemon is configured with an explicit hub address
    When the daemon starts with discovery also enabled
    Then the daemon connects to the explicitly configured hub
    And the daemon ignores the discovered hub

  @story @priority:high
  Scenario: Discovery does not bypass trust for an untrusted hub
    Given a hub is discovered on the network
    Given the discovered hub lacks a valid enrollment token and TLS identity
    When the daemon attempts to enroll with the discovered hub
    Then the daemon refuses to stream to the untrusted hub
    And discovery does not bypass trust verification

  @story @priority:high
  Scenario: Discovery works across the available network providers
    Given discovery providers for the LAN, an overlay network, and a static seed are available
    When a daemon with discovery enabled starts on one of those networks
    Then the daemon locates its hub through whichever provider is reachable
