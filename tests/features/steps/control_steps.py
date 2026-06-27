# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for story 0010 (read-only allow-listed control plane). Each
# step drives the real heimdall-daemon and heimdall-dashboard binaries.
import socket
import subprocess
import time

from behave import given, when, then


def _free_port():
    s = socket.socket()
    s.bind(("127.0.0.1", 0))
    port = s.getsockname()[1]
    s.close()
    return port


def _wait_port(port, timeout=8.0):
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with socket.create_connection(("127.0.0.1", port), 0.2):
                return True
        except OSError:
            time.sleep(0.05)
    return False


def _start_daemon(context, token="sec"):
    port = _free_port()
    logf = open(f"/tmp/heimdall-accept-{port}.log", "w+")
    proc = subprocess.Popen(
        [
            str(context.bin / "heimdall-daemon"),
            "--control-listen", f":{port}",
            "--control-token", token,
            "--interval", "30s",
        ],
        stdout=logf,
        stderr=subprocess.STDOUT,
        cwd=str(context.root),
    )
    context.procs.append(proc)
    context.daemon_port = port
    context.daemon_token = token
    context.daemon_logpath = logf.name
    assert _wait_port(port), "daemon control endpoint did not open"


def _run(context, cmd, token=None):
    token = context.daemon_token if token is None else token
    return subprocess.run(
        [
            str(context.bin / "heimdall-dashboard"),
            "--control", f"localhost:{context.daemon_port}",
            "--token", token,
            "--run", cmd,
        ],
        capture_output=True,
        text=True,
        cwd=str(context.root),
        timeout=15,
    )


def _audit(context):
    time.sleep(0.3)  # let the daemon flush its audit line
    with open(context.daemon_logpath) as f:
        return f.read()


@given(u"the control plane exposes an allow-list of read-only commands on a host")
def step_allowlist_exposed(context):
    _start_daemon(context)


@given(u"the control plane runs commands as the unprivileged daemon user")
def step_runs_unprivileged(context):
    _start_daemon(context)


@given(u"the control plane is enabled on a host")
def step_enabled(context):
    _start_daemon(context)


@when(u"the operator runs an allow-listed query such as listing processes, showing disk usage, or listing files in an allowed directory")
def step_run_allowlisted(context):
    context.result = _run(context, "process.list")


@when(u"the operator attempts to use sudo or run a command that is not on the allow-list")
def step_run_sudo(context):
    context.result = _run(context, "sudo")


@when(u"any control-plane command is invoked")
def step_invoke_any(context):
    context.result = _run(context, "uptime")


@then(u"the host runs the query as the unprivileged user and returns the result")
def step_returns_result(context):
    assert context.result.returncode == 0, context.result.stderr
    assert context.result.stdout.strip() != "", "no result returned"


@then(u"the result is shown in the dashboard")
def step_shown(context):
    assert context.result.stdout.strip() != ""


@then(u"the host refuses the command")
def step_refuses(context):
    assert context.result.returncode != 0, "refused command unexpectedly succeeded"
    err = context.result.stderr.lower()
    assert "refused" in err or "insufficient" in err, context.result.stderr


@then(u"no command is run with elevated privileges")
def step_no_elevation(context):
    assert context.result.stdout.strip() == "", "refused command produced output"
    assert '"decision":"refuse"' in _audit(context)


@then(u"the host records an audit log entry for that invocation")
def step_audit_entry(context):
    assert '"msg":"control audit"' in _audit(context), "no audit entry recorded"


@then(u"the audit entry identifies the command and the requesting operator")
def step_audit_identifies(context):
    text = _audit(context)
    assert '"command":"uptime"' in text, "audit entry missing command"
    assert '"actor":' in text, "audit entry missing operator"
