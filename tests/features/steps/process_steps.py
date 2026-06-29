# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for the v2 process view (top) and privileged commands
# (ADR 0017/0018). Drives the real heimdall-hub, heimdall-daemon, and
# heimdall-cli binaries.
import json
import os
import socket
import subprocess
import tempfile
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


def _cli(context, *args, wait=None):
    cmd = [str(context.bin / "heimdall-cli"), "--hub", f"localhost:{context.hub_port}"]
    if wait:
        cmd += ["--wait", wait]
    cmd += list(args)
    return subprocess.run(cmd, capture_output=True, text=True, cwd=str(context.root), timeout=20)


def _start_hub(context):
    port = _free_port()
    proc = subprocess.Popen(
        [str(context.bin / "heimdall-hub"), "--listen", f":{port}", "--id", "accept"],
        stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, cwd=str(context.root),
    )
    context.procs.append(proc)
    context.hub_port = port
    assert _wait_port(port), "hub did not open its port"


def _start_daemon(context, process=False, allow_commands=False):
    args = [
        str(context.bin / "heimdall-daemon"),
        "--hub", f"localhost:{context.hub_port}",
        "--name", "accept-host",
        "--interval", "1s",
    ]
    if process:
        args += ["--process-interval", "1s"]
    if allow_commands:
        args.append("--allow-commands")
    # Isolate the helper socket to a guaranteed-absent path so the "no helper" case
    # is deterministic regardless of any helper running on the default socket.
    env = dict(os.environ)
    env["HEIMDALL_HELPER_SOCKET"] = tempfile.mktemp(prefix="heimdall-nohelper-", suffix=".sock")
    proc = subprocess.Popen(args, stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, cwd=str(context.root), env=env)
    context.procs.append(proc)
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


@given(u"a hub and a daemon pushing a process table")
def step_hub_daemon_proc(context):
    _start_hub(context)
    _start_daemon(context, process=True)


@when(u"the operator reads the host's process table")
def step_read_top(context):
    # A longer window opens the demand window and catches the pushed table.
    out = _cli(context, "top", "accept-host", wait="6s")
    context.top = json.loads(out.stdout)


@then(u"the process table lists running processes")
def step_top_lists(context):
    procs = context.top.get("processes", [])
    assert len(procs) > 0, f"no processes returned: {context.top}"
    assert any(p.get("command") for p in procs), "process rows have no command"


@given(u"a hub and a command-enabled daemon without a helper")
def step_hub_daemon_cmd_no_helper(context):
    _start_hub(context)
    _start_daemon(context, allow_commands=True)


@when(u"the operator runs a privileged command")
def step_run_privileged(context):
    out = _cli(context, "run", "accept-host", "dmesg")
    context.result = json.loads(out.stdout.strip() or out.stderr.strip())


@then(u"the host reports that the command needs the privileged helper")
def step_needs_helper(context):
    assert context.result.get("status") == "insufficient_permission", context.result
    assert "helper" in context.result.get("stderr", "").lower(), context.result
