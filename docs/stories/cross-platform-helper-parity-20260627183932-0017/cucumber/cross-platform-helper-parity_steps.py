# Step definitions for: Cross-Platform Privileged Metrics Parity
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a privileged helper runs on a Linux host")
def step_a_privileged_helper_runs_on_a_linux_host(context):
    raise NotImplementedError(u"STEP: given a privileged helper runs on a Linux host")


@when(u"the helper reports power and thermal metrics")
def step_the_helper_reports_power_and_thermal_metrics(context):
    raise NotImplementedError(u"STEP: when the helper reports power and thermal metrics")


@then(u"the helper reports CPU package power via RAPL")
def step_the_helper_reports_cpu_package_power_via_rapl(context):
    raise NotImplementedError(u"STEP: then the helper reports CPU package power via RAPL")


@then(u"the helper reports temperatures via hwmon")
def step_the_helper_reports_temperatures_via_hwmon(context):
    raise NotImplementedError(u"STEP: then the helper reports temperatures via hwmon")


@given(u"a Linux host without RAPL support")
def step_a_linux_host_without_rapl_support(context):
    raise NotImplementedError(u"STEP: given a Linux host without RAPL support")


@when(u"the helper reads CPU package power")
def step_the_helper_reads_cpu_package_power(context):
    raise NotImplementedError(u"STEP: when the helper reads CPU package power")


@then(u"the metric reports unavailable or needs-helper")
def step_the_metric_reports_unavailable_or_needs_helper(context):
    raise NotImplementedError(u"STEP: then the metric reports unavailable or needs-helper")


@then(u"the daemon keeps reporting its other metrics")
def step_the_daemon_keeps_reporting_its_other_metrics(context):
    raise NotImplementedError(u"STEP: then the daemon keeps reporting its other metrics")


@given(u"a privileged helper runs on a Windows host")
def step_a_privileged_helper_runs_on_a_windows_host(context):
    raise NotImplementedError(u"STEP: given a privileged helper runs on a Windows host")


@then(u"a privileged path provides power and thermal metrics comparable to Apple Silicon")
def step_a_privileged_path_provides_power_and_thermal_metrics_compara(context):
    raise NotImplementedError(u"STEP: then a privileged path provides power and thermal metrics comparable to Apple Silicon")
