# Step definitions for: Gjallarhorn: Alerting and Notifications
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"an alert rule of cpu.util above 90% for 5m")
def step_an_alert_rule_of_cpu_util_above_90_for_5m(context):
    raise NotImplementedError(u"STEP: given an alert rule of cpu.util above 90% for 5m")


@when(u"a host breaches 90% CPU continuously for 5 minutes")
def step_a_host_breaches_90_cpu_continuously_for_5_minutes(context):
    raise NotImplementedError(u"STEP: when a host breaches 90% CPU continuously for 5 minutes")


@then(u"the rule fires an alert")
def step_the_rule_fires_an_alert(context):
    raise NotImplementedError(u"STEP: then the rule fires an alert")


@then(u"the dashboard shows the alert")
def step_the_dashboard_shows_the_alert(context):
    raise NotImplementedError(u"STEP: then the dashboard shows the alert")


@given(u"an alert rule with a webhook configured")
def step_an_alert_rule_with_a_webhook_configured(context):
    raise NotImplementedError(u"STEP: given an alert rule with a webhook configured")


@when(u"the alert fires and later clears")
def step_the_alert_fires_and_later_clears(context):
    raise NotImplementedError(u"STEP: when the alert fires and later clears")


@then(u"the hub posts a notification to the webhook when it fires")
def step_the_hub_posts_a_notification_to_the_webhook_when_it_fires(context):
    raise NotImplementedError(u"STEP: then the hub posts a notification to the webhook when it fires")


@then(u"the hub posts a notification to the webhook again when it clears")
def step_the_hub_posts_a_notification_to_the_webhook_again_when_it_cl(context):
    raise NotImplementedError(u"STEP: then the hub posts a notification to the webhook again when it clears")


@when(u"a host spikes above 90% briefly and recovers within the for-duration")
def step_a_host_spikes_above_90_briefly_and_recovers_within_the_for_d(context):
    raise NotImplementedError(u"STEP: when a host spikes above 90% briefly and recovers within the for-duration")


@then(u"no alert fires")
def step_no_alert_fires(context):
    raise NotImplementedError(u"STEP: then no alert fires")


@given(u"an alert rule scoped to the tag env=prod")
def step_an_alert_rule_scoped_to_the_tag_env_prod(context):
    raise NotImplementedError(u"STEP: given an alert rule scoped to the tag env=prod")


@when(u"a host tagged env=prod breaches the threshold")
def step_a_host_tagged_env_prod_breaches_the_threshold(context):
    raise NotImplementedError(u"STEP: when a host tagged env=prod breaches the threshold")


@then(u"the alert fires for that host")
def step_the_alert_fires_for_that_host(context):
    raise NotImplementedError(u"STEP: then the alert fires for that host")


@then(u"hosts without the tag are not evaluated by the rule")
def step_hosts_without_the_tag_are_not_evaluated_by_the_rule(context):
    raise NotImplementedError(u"STEP: then hosts without the tag are not evaluated by the rule")
