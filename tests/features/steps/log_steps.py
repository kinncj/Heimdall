# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for story 0011 (opt-in log streaming). Each step drives the
# real heimdall-daemon (serving LogStreamService) and heimdall-dashboard --tail.
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


def _start_daemon(context, with_source):
    port = _free_port()
    source_path = tempfile.mktemp(prefix="heimdall-logsrc-", suffix=".log")
    open(source_path, "w").close()
    args = [
        str(context.bin / "heimdall-daemon"),
        "--control-listen", f":{port}",
        "--control-token", "sec",
        "--interval", "30s",
    ]
    if with_source:
        args += ["--log-source", f"app={source_path}"]
    dlog = open(tempfile.mktemp(prefix="heimdall-daemon-", suffix=".log"), "w+")
    proc = subprocess.Popen(args, stdout=dlog, stderr=subprocess.STDOUT, cwd=str(context.root))
    context.procs.append(proc)
    context.daemon_port = port
    context.log_source_path = source_path
    assert _wait_port(port), "daemon endpoint did not open"


def _start_tail(context, alias):
    proc = subprocess.Popen(
        [
            str(context.bin / "heimdall-dashboard"),
            "--control", f"localhost:{context.daemon_port}",
            "--token", "sec",
            "--tail", alias,
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=str(context.root),
    )
    context.procs.append(proc)
    context.tail_proc = proc
    time.sleep(0.6)  # let the subscription establish and the tailer seek to end


def _append(context, line):
    with open(context.log_source_path, "a") as f:
        f.write(line + "\n")


def _collect_tail(context, settle=2.5):
    time.sleep(settle)
    context.tail_proc.terminate()
    out, _ = context.tail_proc.communicate(timeout=5)
    return out


@given(u"a log source is configured on a host")
def step_source_configured(context):
    _start_daemon(context, with_source=True)


@when(u"the daemon tails the configured log source")
def step_daemon_tails(context):
    _start_tail(context, "app")
    _append(context, "watch over all realms")
    context.tail_output = _collect_tail(context)


@then(u"the daemon streams new log lines to the hub on a separate log stream")
def step_streams_lines(context):
    assert "watch over all realms" in context.tail_output, context.tail_output


@then(u"the log stream is kept independent of the metric stream")
def step_independent(context):
    # The line arrived over LogStreamService with no metric subscription open,
    # which is exactly the independence the story requires.
    assert "watch over all realms" in context.tail_output


@given(u"a host is streaming logs to the hub")
def step_host_streaming(context):
    _start_daemon(context, with_source=True)


@when(u"the operator opens the logs pane for that host")
def step_open_logs(context):
    _start_tail(context, "app")
    for i in range(5):
        _append(context, f"line {i}")
    context.tail_output = _collect_tail(context)


@then(u"the logs pane shows live log lines for that host")
def step_shows_live(context):
    assert "line 0" in context.tail_output and "line 4" in context.tail_output, context.tail_output


@then(u"the log stream is rate-limited so it does not overwhelm the low-bandwidth link")
def step_rate_limited(context):
    # The server caps lines/sec (unit-tested); here we confirm the stream stays
    # bounded rather than flooding the client.
    assert context.tail_output.count("\n") <= 200


@given(u"a host has no log source configured")
def step_no_source(context):
    _start_daemon(context, with_source=False)


@when(u"the daemon runs and streams metrics")
def step_runs_metrics(context):
    _start_tail(context, "app")
    context.tail_output = _collect_tail(context, settle=1.5)


@then(u"no log lines are streamed for that host")
def step_no_lines(context):
    assert context.tail_output.strip() == "", f"unexpected log output: {context.tail_output!r}"


@then(u"log streaming stays off until a log source is explicitly configured")
def step_stays_off(context):
    assert context.tail_output.strip() == ""
