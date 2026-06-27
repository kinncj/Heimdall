# Step definitions for: Durable TSDB Sink and Fleet Restore on Restart
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"the hub is started with a Prometheus-compatible TSDB configured")
def step_the_hub_is_started_with_a_prometheus_compatible_tsdb_configured(context):
    raise NotImplementedError(u"STEP: given the hub is started with a Prometheus-compatible TSDB configured")


@when(u"daemons report metrics to the hub")
def step_daemons_report_metrics_to_the_hub(context):
    raise NotImplementedError(u"STEP: when daemons report metrics to the hub")


@then(u"the hub writes those metrics to the configured TSDB continuously")
def step_the_hub_writes_those_metrics_to_the_configured_tsdb_continuously(context):
    raise NotImplementedError(u"STEP: then the hub writes those metrics to the configured TSDB continuously")


@given(u"the hub previously persisted a fleet to the configured TSDB")
def step_the_hub_previously_persisted_a_fleet_to_the_configured_tsdb(context):
    raise NotImplementedError(u"STEP: given the hub previously persisted a fleet to the configured TSDB")


@when(u"the hub restarts before any daemon reconnects")
def step_the_hub_restarts_before_any_daemon_reconnects(context):
    raise NotImplementedError(u"STEP: when the hub restarts before any daemon reconnects")


@then(u"the hub restores the last-known fleet from the configured TSDB")
def step_the_hub_restores_the_last_known_fleet_from_the_configured_tsdb(context):
    raise NotImplementedError(u"STEP: then the hub restores the last-known fleet from the configured TSDB")


@then(u"offline hosts are restored with their last-seen age")
def step_offline_hosts_are_restored_with_their_last_seen_age(context):
    raise NotImplementedError(u"STEP: then offline hosts are restored with their last-seen age")


@given(u"the hub is started with no TSDB configured")
def step_the_hub_is_started_with_no_tsdb_configured(context):
    raise NotImplementedError(u"STEP: given the hub is started with no TSDB configured")


@when(u"the hub restarts")
def step_the_hub_restarts(context):
    raise NotImplementedError(u"STEP: when the hub restarts")


@then(u"the hub starts with an empty fleet and no restored state")
def step_the_hub_starts_with_an_empty_fleet_and_no_restored_state(context):
    raise NotImplementedError(u"STEP: then the hub starts with an empty fleet and no restored state")


@given(u"the hub restores a fleet from the configured TSDB")
def step_the_hub_restores_a_fleet_from_the_configured_tsdb(context):
    raise NotImplementedError(u"STEP: given the hub restores a fleet from the configured TSDB")


@when(u"the restored fleet is presented before live data resumes")
def step_the_restored_fleet_is_presented_before_live_data_resumes(context):
    raise NotImplementedError(u"STEP: when the restored fleet is presented before live data resumes")


@then(u"info strings and alert state are absent from the restored fleet")
def step_info_strings_and_alert_state_are_absent_from_the_restored_fleet(context):
    raise NotImplementedError(u"STEP: then info strings and alert state are absent from the restored fleet")


@then(u"they reappear once live data resumes")
def step_they_reappear_once_live_data_resumes(context):
    raise NotImplementedError(u"STEP: then they reappear once live data resumes")
