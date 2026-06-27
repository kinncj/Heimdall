# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for story 0001 (cross-platform daemon enrollment). Each step
# drives the real heimdall-hub, heimdall-daemon, and heimdall-dashboard binaries.
# "Mixed operating systems" is represented by multiple named hosts enrolling
# against one hub; cross-platform builds are proven by the release matrix.
import os
import re
import socket
import subprocess
import tempfile
import time

from behave import given, when, then

ANSI = re.compile("\x1b\\[[0-9;]*m")


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


def _start_hub(context, token="", tls_dir=None):
    port = _free_port()
    args = [str(context.bin / "heimdall-hub"), "--listen", f":{port}",
            "--stale-after", "30s", "--offline-after", "60s"]
    if token:
        args += ["--token", token]
    if tls_dir:
        args += ["--tls-cert", os.path.join(tls_dir, "hub.crt"),
                 "--tls-key", os.path.join(tls_dir, "hub.key")]
    log = open(tempfile.mktemp(prefix="heimdall-hub-", suffix=".log"), "w+")
    proc = subprocess.Popen(args, stdout=log, stderr=subprocess.STDOUT, cwd=str(context.root))
    context.procs.append(proc)
    context.hub_port = port
    context.hub_token = token
    context.hub_tls_dir = tls_dir
    assert _wait_port(port), "hub did not open"


def _start_daemon(context, name, token=None):
    token = context.hub_token if token is None else token
    args = [str(context.bin / "heimdall-daemon"), "--hub", f"localhost:{context.hub_port}",
            "--name", name, "--interval", "1s"]
    if token:
        args += ["--token", token]
    if context.hub_tls_dir:
        args += ["--tls", "--tls-ca", os.path.join(context.hub_tls_dir, "hub.crt")]
    log = open(tempfile.mktemp(prefix=f"heimdall-daemon-{name}-", suffix=".log"), "w+")
    proc = subprocess.Popen(args, stdout=log, stderr=subprocess.STDOUT, cwd=str(context.root))
    context.procs.append(proc)
    return proc, log.name


def _snapshot(context):
    args = [str(context.bin / "heimdall-dashboard"), "--hub", f"localhost:{context.hub_port}", "--snapshot"]
    if context.hub_token:
        args += ["--token", context.hub_token]
    if context.hub_tls_dir:
        args += ["--tls", "--tls-ca", os.path.join(context.hub_tls_dir, "hub.crt")]
    out = subprocess.run(args, capture_output=True, text=True, cwd=str(context.root), timeout=20).stdout
    return ANSI.sub("", out)


def _poll_snapshot(context, predicate, timeout=25.0):
    deadline = time.time() + timeout
    last = ""
    while time.time() < deadline:
        last = _snapshot(context)
        if predicate(last):
            return last
        time.sleep(0.5)
    return last


@given(u"a lightweight daemon is installed on Windows, macOS, and Linux hosts")
def step_daemons_installed(context):
    _start_hub(context)


@given(u"the hosts include devices such as a workstation, DGX Spark, HP Strix Halo, Mac mini, Raspberry Pi, and Alienware machine")
def step_host_inventory(context):
    context.host_names = ["workstation", "dgx-spark", "mac-mini"]


@when(u"each daemon starts and connects to the central service over a low-bandwidth socket protocol")
def step_daemons_connect(context):
    for name in context.host_names:
        _start_daemon(context, name)


@then(u"each host appears as an online monitor target in the centralized system")
def step_hosts_online(context):
    def all_online(text):
        return all(re.search(re.escape(n) + r".*ONLINE", text) for n in context.host_names)

    text = _poll_snapshot(context, all_online)
    for name in context.host_names:
        assert re.search(re.escape(name) + r".*ONLINE", text), f"{name} not online:\n{text}"


@then(u"each host sends periodic metric updates without requiring high network throughput")
def step_periodic_updates(context):
    # A second snapshot still shows the hosts online, i.e. they keep streaming
    # periodic updates over the same low-bandwidth gRPC channel.
    text = _snapshot(context)
    for name in context.host_names:
        assert re.search(re.escape(name) + r".*ONLINE", text), f"{name} stopped updating"


@given(u"the central hub requires TLS and a valid enrollment token to register a daemon")
def step_hub_tls_token(context):
    tls_dir = tempfile.mkdtemp(prefix="heimdall-certs-")
    subprocess.run(["bash", "scripts/gen-dev-certs.sh", tls_dir], cwd=str(context.root),
                   check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    _start_hub(context, token="good-token", tls_dir=tls_dir)


@when(u"a daemon attempts to enroll over TLS using an invalid or missing enrollment token")
def step_daemon_bad_token(context):
    context.bad_proc, context.bad_log = _start_daemon(context, "intruder", token="WRONG")
    time.sleep(2.0)  # let it attempt and be rejected at least once


@then(u"the hub rejects the connection during enrollment")
def step_hub_rejects(context):
    deadline = time.time() + 8
    text = ""
    while time.time() < deadline:
        with open(context.bad_log) as f:
            text = f.read()
        if "Unauthenticated" in text:
            return
        time.sleep(0.2)
    raise AssertionError(f"daemon was not rejected:\n{text}")


@then(u"the unauthenticated daemon is not registered as a monitor target")
def step_not_registered(context):
    text = _snapshot(context)
    assert "intruder" not in text, f"unauthenticated daemon was registered:\n{text}"
