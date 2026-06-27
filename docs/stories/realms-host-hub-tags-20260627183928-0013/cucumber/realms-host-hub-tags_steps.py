# Step definitions for: Realms: Host and Hub Tags
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a daemon is started with the tags env=prod and role=db")
def step_a_daemon_is_started_with_the_tags_env_prod_and_role_db(context):
    raise NotImplementedError(u"STEP: given a daemon is started with the tags env=prod and role=db")


@when(u"the daemon's metrics reach the dashboard")
def step_the_daemon_s_metrics_reach_the_dashboard(context):
    raise NotImplementedError(u"STEP: when the daemon's metrics reach the dashboard")


@then(u"the dashboard shows the host carrying the tags env=prod and role=db")
def step_the_dashboard_shows_the_host_carrying_the_tags_env_prod_and(context):
    raise NotImplementedError(u"STEP: then the dashboard shows the host carrying the tags env=prod and role=db")


@given(u"a hub carries the tag region=eu")
def step_a_hub_carries_the_tag_region_eu(context):
    raise NotImplementedError(u"STEP: given a hub carries the tag region=eu")


@given(u"the hub relays several hosts over Bifrost")
def step_the_hub_relays_several_hosts_over_bifrost(context):
    raise NotImplementedError(u"STEP: given the hub relays several hosts over Bifrost")


@when(u"those relayed hosts appear in the dashboard")
def step_those_relayed_hosts_appear_in_the_dashboard(context):
    raise NotImplementedError(u"STEP: when those relayed hosts appear in the dashboard")


@then(u"each relayed host inherits the tag region=eu from its hub")
def step_each_relayed_host_inherits_the_tag_region_eu_from_its_hub(context):
    raise NotImplementedError(u"STEP: then each relayed host inherits the tag region=eu from its hub")


@given(u"a hub carries the tag env=staging")
def step_a_hub_carries_the_tag_env_staging(context):
    raise NotImplementedError(u"STEP: given a hub carries the tag env=staging")


@given(u"a host the hub relays is started with its own tag env=prod")
def step_a_host_the_hub_relays_is_started_with_its_own_tag_env_prod(context):
    raise NotImplementedError(u"STEP: given a host the hub relays is started with its own tag env=prod")


@when(u"the host's tags are resolved")
def step_the_host_s_tags_are_resolved(context):
    raise NotImplementedError(u"STEP: when the host's tags are resolved")


@then(u"the host's own tag env=prod overrides the inherited hub tag")
def step_the_host_s_own_tag_env_prod_overrides_the_inherited_hub_tag(context):
    raise NotImplementedError(u"STEP: then the host's own tag env=prod overrides the inherited hub tag")
