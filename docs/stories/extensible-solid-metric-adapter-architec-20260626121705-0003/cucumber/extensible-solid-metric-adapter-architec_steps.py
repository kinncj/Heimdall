# Step definitions for: Extensible SOLID Metric Adapter Architecture
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"the monitoring system uses a metric adapter interface in both daemon and central processing components")
def step_the_monitoring_system_uses_a_metric_adapter_interface_in_bot(context):
    raise NotImplementedError(u"STEP: given the monitoring system uses a metric adapter interface in both daemon and central processing components")


@when(u"a developer adds a new metric adapter for a new signal")
def step_a_developer_adds_a_new_metric_adapter_for_a_new_signal(context):
    raise NotImplementedError(u"STEP: when a developer adds a new metric adapter for a new signal")


@then(u"the new metric can be collected and displayed without changing existing adapters")
def step_the_new_metric_can_be_collected_and_displayed_without_changi(context):
    raise NotImplementedError(u"STEP: then the new metric can be collected and displayed without changing existing adapters")


@then(u"the design follows SOLID principles with clear responsibilities and replaceable implementations")
def step_the_design_follows_solid_principles_with_clear_responsibilit(context):
    raise NotImplementedError(u"STEP: then the design follows SOLID principles with clear responsibilities and replaceable implementations")


@given(u"multiple metric adapters are active for a host")
def step_multiple_metric_adapters_are_active_for_a_host(context):
    raise NotImplementedError(u"STEP: given multiple metric adapters are active for a host")


@when(u"one adapter fails to collect a metric")
def step_one_adapter_fails_to_collect_a_metric(context):
    raise NotImplementedError(u"STEP: when one adapter fails to collect a metric")


@then(u"the failure is reported for that metric only")
def step_the_failure_is_reported_for_that_metric_only(context):
    raise NotImplementedError(u"STEP: then the failure is reported for that metric only")


@then(u"other adapters continue collecting and sending their metrics normally")
def step_other_adapters_continue_collecting_and_sending_their_metrics(context):
    raise NotImplementedError(u"STEP: then other adapters continue collecting and sending their metrics normally")


@given(u"a metric adapter cannot collect its signal because the metric is unsupported on the host")
def step_a_metric_adapter_cannot_collect_its_signal_because_the_metri(context):
    raise NotImplementedError(u"STEP: given a metric adapter cannot collect its signal because the metric is unsupported on the host")


@given(u"another metric adapter cannot collect its signal because it lacks the required permission")
def step_another_metric_adapter_cannot_collect_its_signal_because_it(context):
    raise NotImplementedError(u"STEP: given another metric adapter cannot collect its signal because it lacks the required permission")


@when(u"the adapters report their status to the dashboard")
def step_the_adapters_report_their_status_to_the_dashboard(context):
    raise NotImplementedError(u"STEP: when the adapters report their status to the dashboard")


@then(u"the unsupported metric is shown as unavailable with an em dash placeholder rather than an error")
def step_the_unsupported_metric_is_shown_as_unavailable_with_an_em_da(context):
    raise NotImplementedError(u"STEP: then the unsupported metric is shown as unavailable with an em dash placeholder rather than an error")


@then(u"the permission-limited metric is shown as needing the privileged helper rather than an error")
def step_the_permission_limited_metric_is_shown_as_needing_the_privil(context):
    raise NotImplementedError(u"STEP: then the permission-limited metric is shown as needing the privileged helper rather than an error")


@then(u"neither status is presented as a collection failure")
def step_neither_status_is_presented_as_a_collection_failure(context):
    raise NotImplementedError(u"STEP: then neither status is presented as a collection failure")


@given(u"the daemon and the hub both depend on the shared metric adapter contract")
def step_the_daemon_and_the_hub_both_depend_on_the_shared_metric_adap(context):
    raise NotImplementedError(u"STEP: given the daemon and the hub both depend on the shared metric adapter contract")


@given(u"the contract is defined by a common versioned schema")
def step_the_contract_is_defined_by_a_common_versioned_schema(context):
    raise NotImplementedError(u"STEP: given the contract is defined by a common versioned schema")


@when(u"a metric adapter is built into both the daemon and the hub")
def step_a_metric_adapter_is_built_into_both_the_daemon_and_the_hub(context):
    raise NotImplementedError(u"STEP: when a metric adapter is built into both the daemon and the hub")


@then(u"both components use the identical adapter interface and metric definitions")
def step_both_components_use_the_identical_adapter_interface_and_metr(context):
    raise NotImplementedError(u"STEP: then both components use the identical adapter interface and metric definitions")


@then(u"a schema version change is applied in one place and consumed by both components")
def step_a_schema_version_change_is_applied_in_one_place_and_consumed(context):
    raise NotImplementedError(u"STEP: then a schema version change is applied in one place and consumed by both components")

