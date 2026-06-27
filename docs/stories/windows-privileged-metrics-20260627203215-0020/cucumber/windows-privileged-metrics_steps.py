# Step definitions for: Windows Privileged Power and Thermal Metrics
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"the helper runs on a Windows host with WMI available")
def step_the_helper_runs_on_a_windows_host_with_wmi_available(context):
    raise NotImplementedError(u"STEP: given the helper runs on a Windows host with WMI available")


@when(u"the helper collects privileged metrics")
def step_the_helper_collects_privileged_metrics(context):
    raise NotImplementedError(u"STEP: when the helper collects privileged metrics")


@then(u"it reports the CPU zone temperature as temp.pkg")
def step_it_reports_the_cpu_zone_temperature_as_temp_pkg(context):
    raise NotImplementedError(u"STEP: then it reports the CPU zone temperature as temp.pkg")


@given(u"the helper runs on a Windows host")
def step_the_helper_runs_on_a_windows_host(context):
    raise NotImplementedError(u"STEP: given the helper runs on a Windows host")


@then(u"it reports CPU package power as unavailable because RAPL is absent")
def step_it_reports_cpu_package_power_as_unavailable_because_rapl_is_absent(context):
    raise NotImplementedError(u"STEP: then it reports CPU package power as unavailable because RAPL is absent")


@given(u"the helper runs on a Windows host where WMI is unavailable")
def step_the_helper_runs_on_a_windows_host_where_wmi_is_unavailable(context):
    raise NotImplementedError(u"STEP: given the helper runs on a Windows host where WMI is unavailable")


@then(u"each affected metric degrades to unavailable instead of failing")
def step_each_affected_metric_degrades_to_unavailable_instead_of_failing(context):
    raise NotImplementedError(u"STEP: then each affected metric degrades to unavailable instead of failing")
