# Step definitions for: Host Context: Locale, Uptime, and OS
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"a daemon is enrolling with the hub")
def step_a_daemon_is_enrolling_with_the_hub(context):
    raise NotImplementedError(u"STEP: given a daemon is enrolling with the hub")


@when(u"the daemon sends its host context")
def step_the_daemon_sends_its_host_context(context):
    raise NotImplementedError(u"STEP: when the daemon sends its host context")


@then(u"the context includes operating system and architecture, hostname, locale and timezone, uptime, and boot time")
def step_the_context_includes_operating_system_and_architecture_hostn(context):
    raise NotImplementedError(u"STEP: then the context includes operating system and architecture, hostname, locale and timezone, uptime, and boot time")


@then(u"the daemon refreshes the host context periodically while connected")
def step_the_daemon_refreshes_the_host_context_periodically_while_con(context):
    raise NotImplementedError(u"STEP: then the daemon refreshes the host context periodically while connected")


@given(u"a host has reported its host context")
def step_a_host_has_reported_its_host_context(context):
    raise NotImplementedError(u"STEP: given a host has reported its host context")


@when(u"the operator opens the host detail view")
def step_the_operator_opens_the_host_detail_view(context):
    raise NotImplementedError(u"STEP: when the operator opens the host detail view")


@then(u"the dashboard shows the operating system and architecture, hostname, locale and timezone, uptime, and boot time")
def step_the_dashboard_shows_the_operating_system_and_architecture_ho(context):
    raise NotImplementedError(u"STEP: then the dashboard shows the operating system and architecture, hostname, locale and timezone, uptime, and boot time")


@given(u"a host is already enrolled and visible in the dashboard")
def step_a_host_is_already_enrolled_and_visible_in_the_dashboard(context):
    raise NotImplementedError(u"STEP: given a host is already enrolled and visible in the dashboard")


@when(u"the host context changes and the daemon sends an update")
def step_the_host_context_changes_and_the_daemon_sends_an_update(context):
    raise NotImplementedError(u"STEP: when the host context changes and the daemon sends an update")


@then(u"the dashboard reflects the updated context for the same host")
def step_the_dashboard_reflects_the_updated_context_for_the_same_host(context):
    raise NotImplementedError(u"STEP: then the dashboard reflects the updated context for the same host")


@then(u"the host is not re-registered or duplicated")
def step_the_host_is_not_re_registered_or_duplicated(context):
    raise NotImplementedError(u"STEP: then the host is not re-registered or duplicated")

