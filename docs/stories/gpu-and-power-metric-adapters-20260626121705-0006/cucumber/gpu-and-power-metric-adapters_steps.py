# Step definitions for: GPU and Power Metric Adapters
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a host has an NVIDIA GPU and the NVML library is available")
def step_a_host_has_an_nvidia_gpu_and_the_nvml_library_is_available(context):
    raise NotImplementedError(u"STEP: given a host has an NVIDIA GPU and the NVML library is available")


@when(u"the GPU adapter collects metrics through NVML")
def step_the_gpu_adapter_collects_metrics_through_nvml(context):
    raise NotImplementedError(u"STEP: when the GPU adapter collects metrics through NVML")


@then(u"it reports GPU utilization, VRAM usage, temperature, and power draw")
def step_it_reports_gpu_utilization_vram_usage_temperature_and_power(context):
    raise NotImplementedError(u"STEP: then it reports GPU utilization, VRAM usage, temperature, and power draw")


@then(u"these values appear for that host in the dashboard")
def step_these_values_appear_for_that_host_in_the_dashboard(context):
    raise NotImplementedError(u"STEP: then these values appear for that host in the dashboard")


@given(u"a host is an Apple Silicon Mac with the privileged helper installed")
def step_a_host_is_an_apple_silicon_mac_with_the_privileged_helper_in(context):
    raise NotImplementedError(u"STEP: given a host is an Apple Silicon Mac with the privileged helper installed")


@when(u"the GPU adapter collects metrics through the Apple Silicon platform path")
def step_the_gpu_adapter_collects_metrics_through_the_apple_silicon_p(context):
    raise NotImplementedError(u"STEP: when the GPU adapter collects metrics through the Apple Silicon platform path")


@then(u"it reports GPU and power metrics for that host")
def step_it_reports_gpu_and_power_metrics_for_that_host(context):
    raise NotImplementedError(u"STEP: then it reports GPU and power metrics for that host")


@then(u"the daemon collects them without requiring elevated privileges itself")
def step_the_daemon_collects_them_without_requiring_elevated_privileg(context):
    raise NotImplementedError(u"STEP: then the daemon collects them without requiring elevated privileges itself")


@given(u"a host has a GPU from an unsupported vendor such as a Raspberry Pi or an unsupported AMD card")
def step_a_host_has_a_gpu_from_an_unsupported_vendor_such_as_a_raspbe(context):
    raise NotImplementedError(u"STEP: given a host has a GPU from an unsupported vendor such as a Raspberry Pi or an unsupported AMD card")


@when(u"the GPU adapter attempts to collect metrics")
def step_the_gpu_adapter_attempts_to_collect_metrics(context):
    raise NotImplementedError(u"STEP: when the GPU adapter attempts to collect metrics")


@then(u"the adapter reports the GPU metrics as unavailable")
def step_the_adapter_reports_the_gpu_metrics_as_unavailable(context):
    raise NotImplementedError(u"STEP: then the adapter reports the GPU metrics as unavailable")


@then(u"the daemon continues collecting other metrics without crashing")
def step_the_daemon_continues_collecting_other_metrics_without_crashi(context):
    raise NotImplementedError(u"STEP: then the daemon continues collecting other metrics without crashing")


@given(u"a host reports a power profile through the power adapter")
def step_a_host_reports_a_power_profile_through_the_power_adapter(context):
    raise NotImplementedError(u"STEP: given a host reports a power profile through the power adapter")


@when(u"the dashboard displays the host power information")
def step_the_dashboard_displays_the_host_power_information(context):
    raise NotImplementedError(u"STEP: when the dashboard displays the host power information")


@then(u"the power profile is shown as read-only information")
def step_the_power_profile_is_shown_as_read_only_information(context):
    raise NotImplementedError(u"STEP: then the power profile is shown as read-only information")


@then(u"the dashboard offers no control to change the power profile")
def step_the_dashboard_offers_no_control_to_change_the_power_profile(context):
    raise NotImplementedError(u"STEP: then the dashboard offers no control to change the power profile")

