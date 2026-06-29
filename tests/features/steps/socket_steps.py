# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions that PROVE the v2 socket model by inspecting the real sockets
# of the running heimdall processes with `ss` (no extra Python deps, so it works
# in behave's pipx venv):
#   - a daemon listens on nothing (no inbound surface);
#   - a daemon holds exactly one outbound connection, to the hub;
#   - an on-demand directive opens no new socket — it rides the existing stream.
# Reuses the hub/daemon startup from process_steps.py (context.hub_proc /
# context.daemon_proc).
import json
import re
import subprocess

from behave import given, when, then

from process_steps import _cli  # noqa: F401 — reuse the CLI driver

_ADDR = re.compile(r"^\[?[0-9A-Fa-f.:]+\]?:\d+$")  # ip:port (v4/v6)


def _ss(*args):
    r = subprocess.run(["ss", "-H", "-n", "-p", *args], capture_output=True, text=True)
    assert r.returncode == 0, f"ss {' '.join(args)} failed ({r.returncode}): {r.stderr.strip()}"
    return r.stdout.splitlines()


def _for_pid(lines, pid):
    return [ln for ln in lines if f"pid={pid}," in ln]


def _addr_ports(line):
    return [t for t in line.split() if _ADDR.match(t)]


def _peer_port(line):
    aps = _addr_ports(line)
    return int(aps[-1].rsplit(":", 1)[1]) if aps else None


def _inet_listeners(pid):
    return _for_pid(_ss("-lt"), pid)


def _unix_listeners(pid):
    return _for_pid(_ss("-lx"), pid)


def _inet_established(pid):
    return _for_pid(_ss("-t", "state", "established"), pid)


def _fingerprint(pid):
    # The daemon's live endpoints: a new socket (inet or unix) would change this.
    peers = tuple(sorted(_addr_ports(ln)[-1] for ln in _inet_established(pid) if _addr_ports(ln)))
    unix = len(_for_pid(_ss("-x", "state", "connected"), pid))
    return peers, unix


@then(u"the daemon process is listening on no network port")
def step_daemon_no_inet_listen(context):
    lis = _inet_listeners(context.daemon_proc.pid)
    assert not lis, "daemon is listening — daemons must never listen:\n" + "\n".join(lis)


@then(u"the daemon process is listening on no unix socket")
def step_daemon_no_unix_listen(context):
    lis = _unix_listeners(context.daemon_proc.pid)
    assert not lis, "daemon holds a unix listener:\n" + "\n".join(lis)


@then(u"the hub is the only one of our processes holding a listening network port")
def step_hub_sole_listener(context):
    for proc in context.procs:
        if not proc.pid or proc.poll() is not None:
            continue
        listening = bool(_inet_listeners(proc.pid))
        if proc.pid == context.hub_proc.pid:
            assert listening, "the hub must hold a listening port"
        else:
            assert not listening, f"pid {proc.pid} listens but is not the hub"


@then(u"the daemon has exactly one established network connection, to the hub")
def step_daemon_one_conn(context):
    est = _inet_established(context.daemon_proc.pid)
    to_hub = [ln for ln in est if _peer_port(ln) == context.hub_port]
    assert len(est) == 1 and len(to_hub) == 1, (
        f"want exactly one daemon→hub connection on :{context.hub_port}, got:\n" + "\n".join(est))


@given(u"the daemon's established connections are recorded")
def step_record_conns(context):
    context.daemon_conns_before = _fingerprint(context.daemon_proc.pid)


@when(u"the operator runs an allow-listed command on the host")
def step_run_allowlisted(context):
    # uptime is read-only and unprivileged, so the daemon runs it locally and
    # returns the result down the existing stream — no helper, no new socket.
    out = _cli(context, "run", "accept-host", "uptime")
    context.result = json.loads(out.stdout.strip() or out.stderr.strip())


@then(u"the command result returns")
def step_cmd_returned(context):
    assert context.result.get("status") == "ok", context.result


@then(u"the daemon opened no new connection")
def step_no_new_conn(context):
    after = _fingerprint(context.daemon_proc.pid)
    assert after == context.daemon_conns_before, (
        "daemon connection set changed after the command — a new socket was opened.\n"
        f"before={context.daemon_conns_before}\nafter ={after}")
