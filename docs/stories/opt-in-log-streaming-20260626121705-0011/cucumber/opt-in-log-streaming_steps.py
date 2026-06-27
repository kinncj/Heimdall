# Step definitions for: Opt-In Log Streaming
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a log source is configured on a host")
def step_a_log_source_is_configured_on_a_host(context):
    raise NotImplementedError(u"STEP: given a log source is configured on a host")


@when(u"the daemon tails the configured log source")
def step_the_daemon_tails_the_configured_log_source(context):
    raise NotImplementedError(u"STEP: when the daemon tails the configured log source")


@then(u"the daemon streams new log lines to the hub on a separate log stream")
def step_the_daemon_streams_new_log_lines_to_the_hub_on_a_separate_lo(context):
    raise NotImplementedError(u"STEP: then the daemon streams new log lines to the hub on a separate log stream")


@then(u"the log stream is kept independent of the metric stream")
def step_the_log_stream_is_kept_independent_of_the_metric_stream(context):
    raise NotImplementedError(u"STEP: then the log stream is kept independent of the metric stream")


@given(u"a host is streaming logs to the hub")
def step_a_host_is_streaming_logs_to_the_hub(context):
    raise NotImplementedError(u"STEP: given a host is streaming logs to the hub")


@when(u"the operator opens the logs pane for that host")
def step_the_operator_opens_the_logs_pane_for_that_host(context):
    raise NotImplementedError(u"STEP: when the operator opens the logs pane for that host")


@then(u"the logs pane shows live log lines for that host")
def step_the_logs_pane_shows_live_log_lines_for_that_host(context):
    raise NotImplementedError(u"STEP: then the logs pane shows live log lines for that host")


@then(u"the log stream is rate-limited so it does not overwhelm the low-bandwidth link")
def step_the_log_stream_is_rate_limited_so_it_does_not_overwhelm_the(context):
    raise NotImplementedError(u"STEP: then the log stream is rate-limited so it does not overwhelm the low-bandwidth link")


@given(u"a host has no log source configured")
def step_a_host_has_no_log_source_configured(context):
    raise NotImplementedError(u"STEP: given a host has no log source configured")


@when(u"the daemon runs and streams metrics")
def step_the_daemon_runs_and_streams_metrics(context):
    raise NotImplementedError(u"STEP: when the daemon runs and streams metrics")


@then(u"no log lines are streamed for that host")
def step_no_log_lines_are_streamed_for_that_host(context):
    raise NotImplementedError(u"STEP: then no log lines are streamed for that host")


@then(u"log streaming stays off until a log source is explicitly configured")
def step_log_streaming_stays_off_until_a_log_source_is_explicitly_con(context):
    raise NotImplementedError(u"STEP: then log streaming stays off until a log source is explicitly configured")

