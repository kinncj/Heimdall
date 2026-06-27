# Step definitions for: Unprivileged Terminal Control Plane
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


@given(u"the control plane exposes an allow-list of read-only commands on a host")
def step_the_control_plane_exposes_an_allow_list_of_read_only_command(context):
    raise NotImplementedError(u"STEP: given the control plane exposes an allow-list of read-only commands on a host")


@when(u"the operator runs an allow-listed query such as listing processes, showing disk usage, or listing files in an allowed directory")
def step_the_operator_runs_an_allow_listed_query_such_as_listing_proc(context):
    raise NotImplementedError(u"STEP: when the operator runs an allow-listed query such as listing processes, showing disk usage, or listing files in an allowed directory")


@then(u"the host runs the query as the unprivileged user and returns the result")
def step_the_host_runs_the_query_as_the_unprivileged_user_and_returns(context):
    raise NotImplementedError(u"STEP: then the host runs the query as the unprivileged user and returns the result")


@then(u"the result is shown in the dashboard")
def step_the_result_is_shown_in_the_dashboard(context):
    raise NotImplementedError(u"STEP: then the result is shown in the dashboard")


@given(u"the control plane runs commands as the unprivileged daemon user")
def step_the_control_plane_runs_commands_as_the_unprivileged_daemon_u(context):
    raise NotImplementedError(u"STEP: given the control plane runs commands as the unprivileged daemon user")


@when(u"the operator attempts to use sudo or run a command that is not on the allow-list")
def step_the_operator_attempts_to_use_sudo_or_run_a_command_that_is_n(context):
    raise NotImplementedError(u"STEP: when the operator attempts to use sudo or run a command that is not on the allow-list")


@then(u"the host refuses the command")
def step_the_host_refuses_the_command(context):
    raise NotImplementedError(u"STEP: then the host refuses the command")


@then(u"no command is run with elevated privileges")
def step_no_command_is_run_with_elevated_privileges(context):
    raise NotImplementedError(u"STEP: then no command is run with elevated privileges")


@given(u"the control plane is enabled on a host")
def step_the_control_plane_is_enabled_on_a_host(context):
    raise NotImplementedError(u"STEP: given the control plane is enabled on a host")


@when(u"any control-plane command is invoked")
def step_any_control_plane_command_is_invoked(context):
    raise NotImplementedError(u"STEP: when any control-plane command is invoked")


@then(u"the host records an audit log entry for that invocation")
def step_the_host_records_an_audit_log_entry_for_that_invocation(context):
    raise NotImplementedError(u"STEP: then the host records an audit log entry for that invocation")


@then(u"the audit entry identifies the command and the requesting operator")
def step_the_audit_entry_identifies_the_command_and_the_requesting_op(context):
    raise NotImplementedError(u"STEP: then the audit entry identifies the command and the requesting operator")

