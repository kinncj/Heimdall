# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for the v2 hub-mediated allow-listed command plane (ADR 0018).
# Each step drives the real heimdall-hub, heimdall-daemon, and heimdall-cli
# binaries end to end — no mocks, no daemon listener.
import json
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


def _start_hub(context):
    port = _free_port()
    proc = subprocess.Popen(
        [str(context.bin / "heimdall-hub"), "--listen", f":{port}", "--id", "accept"],
        stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, cwd=str(context.root),
    )
    context.procs.append(proc)
    context.hub_port = port
    assert _wait_port(port), "hub did not open its port"


def _start_daemon(context, allow_commands=True):
    args = [
        str(context.bin / "heimdall-daemon"),
        "--hub", f"localhost:{context.hub_port}",
        "--name", "accept-host",
        "--interval", "30s",
    ]
    if allow_commands:
        args.append("--allow-commands")
    logf = open(f"/tmp/heimdall-accept-daemon-{context.hub_port}.log", "w+")
    proc = subprocess.Popen(args, stdout=logf, stderr=subprocess.STDOUT, cwd=str(context.root))
    context.procs.append(proc)
    context.daemon_logpath = logf.name
    # wait for the host to register with the hub
    deadline = time.time() + 10
    while time.time() < deadline:
        out = _cli(context, "hosts")
        try:
            if any(h.get("id") == "accept-host" for h in json.loads(out.stdout)):
                return
        except (json.JSONDecodeError, AttributeError):
            pass
        time.sleep(0.2)
    raise AssertionError("daemon did not register with the hub")


def _cli(context, *args):
    return subprocess.run(
        [str(context.bin / "heimdall-cli"), "--hub", f"localhost:{context.hub_port}", *args],
        capture_output=True, text=True, cwd=str(context.root), timeout=20,
    )


def _run(context, cmd):
    return _cli(context, "run", "accept-host", cmd)


def _audit(context):
    time.sleep(0.3)
    with open(context.daemon_logpath) as f:
        return f.read()


@given(u"a hub and a command-enabled daemon")
def step_hub_and_daemon(context):
    _start_hub(context)
    _start_daemon(context, allow_commands=True)


@given(u"a hub and a daemon with commands disabled")
def step_hub_and_daemon_disabled(context):
    _start_hub(context)
    _start_daemon(context, allow_commands=False)


@when(u"the operator runs an allow-listed command via the CLI")
def step_run_allowlisted(context):
    context.result = _run(context, "disk.df")


@when(u"the operator runs a command that is not on the allow-list")
def step_run_denied(context):
    context.result = _run(context, "rm.rf")


def _result_json(context):
    # `run` prints the JSON result on stdout; refusals from the hub print on stderr.
    out = context.result.stdout.strip() or context.result.stderr.strip()
    return json.loads(out)


@then(u"the command runs on the host and the JSON result is returned")
def step_returns_result(context):
    res = _result_json(context)
    assert res.get("status") == "ok", res
    assert res.get("stdout", "").strip() != "", "no output returned"


@then(u"the host refuses it with insufficient_permission and runs nothing")
def step_refuses(context):
    res = _result_json(context)
    assert res.get("status") == "insufficient_permission", res
    assert res.get("stdout", "") == "", "refused command produced output"


@then(u"the host refuses it because commands are disabled")
def step_refuses_disabled(context):
    res = _result_json(context)
    assert res.get("status") == "insufficient_permission", res
    assert "disabled" in res.get("stderr", "").lower(), res


@then(u"the daemon records an audit entry naming the command and the operator")
def step_audit(context):
    text = _audit(context)
    assert '"msg":"control command"' in text, "no audit entry recorded"
    assert '"cmd":"disk.df"' in text, "audit entry missing the command"
    assert '"actor":' in text, "audit entry missing the operator"
