# Step definitions for: Ratatoskr: Daemon Auto-Discovers Its Hub
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a hub is advertising itself on the local network")
def step_a_hub_is_advertising_itself_on_the_local_network(context):
    raise NotImplementedError(u"STEP: given a hub is advertising itself on the local network")


@given(u"a daemon has discovery enabled and no hub is configured")
def step_a_daemon_has_discovery_enabled_and_no_hub_is_configured(context):
    raise NotImplementedError(u"STEP: given a daemon has discovery enabled and no hub is configured")


@when(u"the daemon starts")
def step_the_daemon_starts(context):
    raise NotImplementedError(u"STEP: when the daemon starts")


@then(u"the daemon discovers the advertised hub")
def step_the_daemon_discovers_the_advertised_hub(context):
    raise NotImplementedError(u"STEP: then the daemon discovers the advertised hub")


@then(u"the daemon streams its metrics to that hub without a hub being hand-configured")
def step_the_daemon_streams_its_metrics_to_that_hub_without_a_hub_bei(context):
    raise NotImplementedError(u"STEP: then the daemon streams its metrics to that hub without a hub being hand-configured")


@given(u"a daemon is configured with an explicit hub address")
def step_a_daemon_is_configured_with_an_explicit_hub_address(context):
    raise NotImplementedError(u"STEP: given a daemon is configured with an explicit hub address")


@when(u"the daemon starts with discovery also enabled")
def step_the_daemon_starts_with_discovery_also_enabled(context):
    raise NotImplementedError(u"STEP: when the daemon starts with discovery also enabled")


@then(u"the daemon connects to the explicitly configured hub")
def step_the_daemon_connects_to_the_explicitly_configured_hub(context):
    raise NotImplementedError(u"STEP: then the daemon connects to the explicitly configured hub")


@then(u"the daemon ignores the discovered hub")
def step_the_daemon_ignores_the_discovered_hub(context):
    raise NotImplementedError(u"STEP: then the daemon ignores the discovered hub")


@given(u"a hub is discovered on the network")
def step_a_hub_is_discovered_on_the_network(context):
    raise NotImplementedError(u"STEP: given a hub is discovered on the network")


@given(u"the discovered hub lacks a valid enrollment token and TLS identity")
def step_the_discovered_hub_lacks_a_valid_enrollment_token_and_tls_id(context):
    raise NotImplementedError(u"STEP: given the discovered hub lacks a valid enrollment token and TLS identity")


@when(u"the daemon attempts to enroll with the discovered hub")
def step_the_daemon_attempts_to_enroll_with_the_discovered_hub(context):
    raise NotImplementedError(u"STEP: when the daemon attempts to enroll with the discovered hub")


@then(u"the daemon refuses to stream to the untrusted hub")
def step_the_daemon_refuses_to_stream_to_the_untrusted_hub(context):
    raise NotImplementedError(u"STEP: then the daemon refuses to stream to the untrusted hub")


@then(u"discovery does not bypass trust verification")
def step_discovery_does_not_bypass_trust_verification(context):
    raise NotImplementedError(u"STEP: then discovery does not bypass trust verification")


@given(u"discovery providers for the LAN, an overlay network, and a static seed are available")
def step_discovery_providers_for_the_lan_an_overlay_network_and_a_sta(context):
    raise NotImplementedError(u"STEP: given discovery providers for the LAN, an overlay network, and a static seed are available")


@when(u"a daemon with discovery enabled starts on one of those networks")
def step_a_daemon_with_discovery_enabled_starts_on_one_of_those_netwo(context):
    raise NotImplementedError(u"STEP: when a daemon with discovery enabled starts on one of those networks")


@then(u"the daemon locates its hub through whichever provider is reachable")
def step_the_daemon_locates_its_hub_through_whichever_provider_is_rea(context):
    raise NotImplementedError(u"STEP: then the daemon locates its hub through whichever provider is reachable")
