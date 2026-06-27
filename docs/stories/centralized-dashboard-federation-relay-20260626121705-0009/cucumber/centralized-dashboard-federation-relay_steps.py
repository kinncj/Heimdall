# Step definitions for: Centralized Dashboard Federation Relay
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a local hub is collecting metrics from its hosts")
def step_a_local_hub_is_collecting_metrics_from_its_hosts(context):
    raise NotImplementedError(u"STEP: given a local hub is collecting metrics from its hosts")


@given(u"a parent cloud hub is configured as an upstream")
def step_a_parent_cloud_hub_is_configured_as_an_upstream(context):
    raise NotImplementedError(u"STEP: given a parent cloud hub is configured as an upstream")


@when(u"the local hub connects to the parent hub as a subscriber")
def step_the_local_hub_connects_to_the_parent_hub_as_a_subscriber(context):
    raise NotImplementedError(u"STEP: when the local hub connects to the parent hub as a subscriber")


@then(u"the parent hub receives the relayed metric stream for the local hosts")
def step_the_parent_hub_receives_the_relayed_metric_stream_for_the_lo(context):
    raise NotImplementedError(u"STEP: then the parent hub receives the relayed metric stream for the local hosts")


@then(u"operators on the parent hub can monitor those hosts")
def step_operators_on_the_parent_hub_can_monitor_those_hosts(context):
    raise NotImplementedError(u"STEP: then operators on the parent hub can monitor those hosts")


@given(u"a hub is publishing host metrics")
def step_a_hub_is_publishing_host_metrics(context):
    raise NotImplementedError(u"STEP: given a hub is publishing host metrics")


@when(u"several dashboards subscribe to the same hub at the same time")
def step_several_dashboards_subscribe_to_the_same_hub_at_the_same_tim(context):
    raise NotImplementedError(u"STEP: when several dashboards subscribe to the same hub at the same time")


@then(u"every subscribed dashboard receives the host metrics")
def step_every_subscribed_dashboard_receives_the_host_metrics(context):
    raise NotImplementedError(u"STEP: then every subscribed dashboard receives the host metrics")


@then(u"all dashboards show consistent data for the same hosts")
def step_all_dashboards_show_consistent_data_for_the_same_hosts(context):
    raise NotImplementedError(u"STEP: then all dashboards show consistent data for the same hosts")


@given(u"a local hub is relaying metrics to an authenticated parent hub")
def step_a_local_hub_is_relaying_metrics_to_an_authenticated_parent_h(context):
    raise NotImplementedError(u"STEP: given a local hub is relaying metrics to an authenticated parent hub")


@when(u"the cross-hub link drops and later reconnects")
def step_the_cross_hub_link_drops_and_later_reconnects(context):
    raise NotImplementedError(u"STEP: when the cross-hub link drops and later reconnects")


@then(u"the parent hub re-authenticates the local hub before accepting data")
def step_the_parent_hub_re_authenticates_the_local_hub_before_accepti(context):
    raise NotImplementedError(u"STEP: then the parent hub re-authenticates the local hub before accepting data")


@then(u"host data resumes without duplication or corruption")
def step_host_data_resumes_without_duplication_or_corruption(context):
    raise NotImplementedError(u"STEP: then host data resumes without duplication or corruption")


@then(u"the relay does not create a loop between hubs")
def step_the_relay_does_not_create_a_loop_between_hubs(context):
    raise NotImplementedError(u"STEP: then the relay does not create a loop between hubs")

