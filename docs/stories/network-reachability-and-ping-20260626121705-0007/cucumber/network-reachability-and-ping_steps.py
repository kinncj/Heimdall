# Step definitions for: Network Reachability and Ping
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a daemon is configured with a reachability target")
def step_a_daemon_is_configured_with_a_reachability_target(context):
    raise NotImplementedError(u"STEP: given a daemon is configured with a reachability target")


@when(u"the daemon probes the configured target and the public internet")
def step_the_daemon_probes_the_configured_target_and_the_public_inter(context):
    raise NotImplementedError(u"STEP: when the daemon probes the configured target and the public internet")


@then(u"it reports the measured latency to the configured target")
def step_it_reports_the_measured_latency_to_the_configured_target(context):
    raise NotImplementedError(u"STEP: then it reports the measured latency to the configured target")


@then(u"it reports internet reachability as up or down")
def step_it_reports_internet_reachability_as_up_or_down(context):
    raise NotImplementedError(u"STEP: then it reports internet reachability as up or down")


@given(u"a host is reporting reachability and latency metrics")
def step_a_host_is_reporting_reachability_and_latency_metrics(context):
    raise NotImplementedError(u"STEP: given a host is reporting reachability and latency metrics")


@when(u"the operator views that host in the dashboard")
def step_the_operator_views_that_host_in_the_dashboard(context):
    raise NotImplementedError(u"STEP: when the operator views that host in the dashboard")


@then(u"the dashboard shows the reachability state with both a symbol and text")
def step_the_dashboard_shows_the_reachability_state_with_both_a_symbo(context):
    raise NotImplementedError(u"STEP: then the dashboard shows the reachability state with both a symbol and text")


@then(u"the dashboard shows a latency trend over recent history")
def step_the_dashboard_shows_a_latency_trend_over_recent_history(context):
    raise NotImplementedError(u"STEP: then the dashboard shows a latency trend over recent history")


@given(u"a host is online and streaming metrics")
def step_a_host_is_online_and_streaming_metrics(context):
    raise NotImplementedError(u"STEP: given a host is online and streaming metrics")


@when(u"the reachability probe fails for that host")
def step_the_reachability_probe_fails_for_that_host(context):
    raise NotImplementedError(u"STEP: when the reachability probe fails for that host")


@then(u"the host remains shown as online")
def step_the_host_remains_shown_as_online(context):
    raise NotImplementedError(u"STEP: then the host remains shown as online")


@then(u"only the reachability metric is shown in an error state")
def step_only_the_reachability_metric_is_shown_in_an_error_state(context):
    raise NotImplementedError(u"STEP: then only the reachability metric is shown in an error state")


@then(u"the other metrics for that host keep updating normally")
def step_the_other_metrics_for_that_host_keep_updating_normally(context):
    raise NotImplementedError(u"STEP: then the other metrics for that host keep updating normally")

