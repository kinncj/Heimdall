# Step definitions for: Optional Privileged Metrics Helper
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"the optional privileged helper is installed on a host")
def step_the_optional_privileged_helper_is_installed_on_a_host(context):
    raise NotImplementedError(u"STEP: given the optional privileged helper is installed on a host")


@given(u"the metrics daemon is running as an unprivileged user")
def step_the_metrics_daemon_is_running_as_an_unprivileged_user(context):
    raise NotImplementedError(u"STEP: given the metrics daemon is running as an unprivileged user")


@when(u"the daemon requests power and full thermal metrics through the helper")
def step_the_daemon_requests_power_and_full_thermal_metrics_through_t(context):
    raise NotImplementedError(u"STEP: when the daemon requests power and full thermal metrics through the helper")


@then(u"the power and full thermal metrics are collected and reported")
def step_the_power_and_full_thermal_metrics_are_collected_and_reporte(context):
    raise NotImplementedError(u"STEP: then the power and full thermal metrics are collected and reported")


@then(u"the daemon continues to run without elevated privileges")
def step_the_daemon_continues_to_run_without_elevated_privileges(context):
    raise NotImplementedError(u"STEP: then the daemon continues to run without elevated privileges")


@given(u"the optional privileged helper is not installed on a host")
def step_the_optional_privileged_helper_is_not_installed_on_a_host(context):
    raise NotImplementedError(u"STEP: given the optional privileged helper is not installed on a host")


@when(u"the power and full thermal adapters attempt to collect their metrics")
def step_the_power_and_full_thermal_adapters_attempt_to_collect_their(context):
    raise NotImplementedError(u"STEP: when the power and full thermal adapters attempt to collect their metrics")


@then(u"those adapters report insufficient permission rather than an error")
def step_those_adapters_report_insufficient_permission_rather_than_an(context):
    raise NotImplementedError(u"STEP: then those adapters report insufficient permission rather than an error")


@then(u"the dashboard shows a needs-helper affordance for those metrics")
def step_the_dashboard_shows_a_needs_helper_affordance_for_those_metr(context):
    raise NotImplementedError(u"STEP: then the dashboard shows a needs-helper affordance for those metrics")


@then(u"the daemon keeps running without crashing")
def step_the_daemon_keeps_running_without_crashing(context):
    raise NotImplementedError(u"STEP: then the daemon keeps running without crashing")


@given(u"the privileged helper runs as its own privileged unit on the host")
def step_the_privileged_helper_runs_as_its_own_privileged_unit_on_the(context):
    raise NotImplementedError(u"STEP: given the privileged helper runs as its own privileged unit on the host")


@given(u"the helper exposes privileged metrics to the daemon over a local socket")
def step_the_helper_exposes_privileged_metrics_to_the_daemon_over_a_l(context):
    raise NotImplementedError(u"STEP: given the helper exposes privileged metrics to the daemon over a local socket")


@when(u"the unprivileged daemon collects privileged metrics")
def step_the_unprivileged_daemon_collects_privileged_metrics(context):
    raise NotImplementedError(u"STEP: when the unprivileged daemon collects privileged metrics")


@then(u"the daemon obtains the values through the local socket")
def step_the_daemon_obtains_the_values_through_the_local_socket(context):
    raise NotImplementedError(u"STEP: then the daemon obtains the values through the local socket")


@then(u"the daemon never invokes sudo or runs as root")
def step_the_daemon_never_invokes_sudo_or_runs_as_root(context):
    raise NotImplementedError(u"STEP: then the daemon never invokes sudo or runs as root")

