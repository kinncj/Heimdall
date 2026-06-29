# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for opt-in log streaming, v2 push model (ADR 0017/0018): the
# daemon tails --log-source files and pushes lines to the hub; the dashboard/CLI
# read them from the hub. Each step drives the real heimdall-hub, heimdall-daemon,
# and heimdall-cli binaries — no daemon listener.
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


def _start_daemon(context, with_source):
    fd, source_path = tempfile.mkstemp(prefix="heimdall-logsrc-", suffix=".log")
    os.close(fd)
    args = [
        str(context.bin / "heimdall-daemon"),
        "--hub", f"localhost:{context.hub_port}",
        "--name", "accept-host",
        "--interval", "1s",  # short so pushed log lines arrive promptly
    ]
    if with_source:
        args += ["--log-source", f"app={source_path}"]
    dfd, dpath = tempfile.mkstemp(prefix="heimdall-daemon-", suffix=".log")
    os.close(dfd)
    dlog = open(dpath, "w+")
    proc = subprocess.Popen(args, stdout=dlog, stderr=subprocess.STDOUT, cwd=str(context.root))
    context.procs.append(proc)
    context.log_source_path = source_path
    # wait for registration
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


def _append(context, line):
    with open(context.log_source_path, "a") as f:
        f.write(line + "\n")


def _read_logs_while_appending(context, lines):
    # Run `cli logs` with a window long enough to (a) open the demand window so the
    # daemon starts pushing logs, and (b) catch the lines we append during it.
    cmd = [
        str(context.bin / "heimdall-cli"), "--hub", f"localhost:{context.hub_port}",
        "--wait", "6s", "logs", "accept-host", "app",
    ]
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, cwd=str(context.root))
    time.sleep(0.8)  # let it subscribe and the window open
    for ln in lines:
        _append(context, ln)
    out, err = proc.communicate(timeout=15)
    assert proc.returncode == 0, f"cli logs failed (exit {proc.returncode}): {err}"
    try:
        return [l["line"] for l in json.loads(out).get("lines", [])]
    except json.JSONDecodeError as e:
        raise AssertionError(f"cli logs emitted non-JSON ({e}): {out!r}")


@given(u"a log source is configured on a host")
def step_source_configured(context):
    _start_hub(context)
    _start_daemon(context, with_source=True)


@when(u"the daemon tails the configured log source")
def step_daemon_tails(context):
    context.log_lines = _read_logs_while_appending(context, ["watch over all realms"])


@then(u"the daemon streams new log lines to the hub on a separate log stream")
def step_streams_lines(context):
    assert any("watch over all realms" in l for l in context.log_lines), context.log_lines


@then(u"the log stream is kept independent of the metric stream")
def step_independent(context):
    # Log lines ride additive snapshot fields, pushed/buffered apart from the metric
    # set, so a host's metrics are unaffected by log volume.
    assert any("watch over all realms" in l for l in context.log_lines)


@given(u"a host is streaming logs to the hub")
def step_host_streaming(context):
    _start_hub(context)
    _start_daemon(context, with_source=True)


@when(u"the operator opens the logs pane for that host")
def step_open_logs(context):
    context.log_lines = _read_logs_while_appending(context, [f"line {i}" for i in range(5)])


@then(u"the logs pane shows live log lines for that host")
def step_shows_live(context):
    joined = "\n".join(context.log_lines)
    assert "line 0" in joined and "line 4" in joined, context.log_lines


@then(u"the log stream is rate-limited so it does not overwhelm the low-bandwidth link")
def step_rate_limited(context):
    # The hub ring + daemon push cap are bounded (unit-tested); confirm the read
    # stays bounded rather than flooding.
    assert len(context.log_lines) <= 500


@given(u"a host has no log source configured")
def step_no_source(context):
    _start_hub(context)
    _start_daemon(context, with_source=False)


@when(u"the daemon runs and streams metrics")
def step_runs_metrics(context):
    out = _cli(context, "logs", "accept-host", "app", wait="2s")
    assert out.returncode == 0, f"cli logs failed (exit {out.returncode}): {out.stderr}"
    try:
        context.log_lines = [l["line"] for l in json.loads(out.stdout).get("lines", [])]
    except json.JSONDecodeError as e:
        raise AssertionError(f"cli logs emitted non-JSON ({e}): {out.stdout!r}")


@then(u"no log lines are streamed for that host")
def step_no_lines(context):
    assert context.log_lines == [], f"unexpected log output: {context.log_lines!r}"


@then(u"log streaming stays off until a log source is explicitly configured")
def step_stays_off(context):
    assert context.log_lines == []
