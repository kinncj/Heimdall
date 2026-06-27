# Step definitions for: Mimir: Metrics History and OpenMetrics Export
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a hub is collecting fleet metrics")
def step_a_hub_is_collecting_fleet_metrics(context):
    raise NotImplementedError(u"STEP: given a hub is collecting fleet metrics")


@when(u"a Prometheus scraper reads the hub's /metrics endpoint")
def step_a_prometheus_scraper_reads_the_hub_s_metrics_endpoint(context):
    raise NotImplementedError(u"STEP: when a Prometheus scraper reads the hub's /metrics endpoint")


@then(u"the scraper receives the metrics in OpenMetrics format")
def step_the_scraper_receives_the_metrics_in_openmetrics_format(context):
    raise NotImplementedError(u"STEP: then the scraper receives the metrics in OpenMetrics format")


@given(u"the hub exports metrics in OpenMetrics format")
def step_the_hub_exports_metrics_in_openmetrics_format(context):
    raise NotImplementedError(u"STEP: given the hub exports metrics in OpenMetrics format")


@when(u"a series is exported for a host")
def step_a_series_is_exported_for_a_host(context):
    raise NotImplementedError(u"STEP: when a series is exported for a host")


@then(u"the series carries the host id, origin hub, and tags as labels")
def step_the_series_carries_the_host_id_origin_hub_and_tags_as_labels(context):
    raise NotImplementedError(u"STEP: then the series carries the host id, origin hub, and tags as labels")


@given(u"the hub retains recent samples in a bounded in-memory window")
def step_the_hub_retains_recent_samples_in_a_bounded_in_memory_window(context):
    raise NotImplementedError(u"STEP: given the hub retains recent samples in a bounded in-memory window")


@when(u"an operator requests recent trends for a host")
def step_an_operator_requests_recent_trends_for_a_host(context):
    raise NotImplementedError(u"STEP: when an operator requests recent trends for a host")


@then(u"the hub serves the retained samples as a short-range trend")
def step_the_hub_serves_the_retained_samples_as_a_short_range_trend(context):
    raise NotImplementedError(u"STEP: then the hub serves the retained samples as a short-range trend")


@given(u"the hub holds recent samples only in memory")
def step_the_hub_holds_recent_samples_only_in_memory(context):
    raise NotImplementedError(u"STEP: given the hub holds recent samples only in memory")


@when(u"the hub restarts")
def step_the_hub_restarts(context):
    raise NotImplementedError(u"STEP: when the hub restarts")


@then(u"the prior history is gone")
def step_the_prior_history_is_gone(context):
    raise NotImplementedError(u"STEP: then the prior history is gone")


@then(u"this loss is documented as acceptable")
def step_this_loss_is_documented_as_acceptable(context):
    raise NotImplementedError(u"STEP: then this loss is documented as acceptable")
