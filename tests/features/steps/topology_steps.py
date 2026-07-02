# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Step definitions for per-core topology and per-cluster CPU power (story 0028).
# Drives the real heimdall-daemon --once and asserts the metric contract on
# whatever hardware the suite runs on: hybrid hosts must report a labelled
# topology, uniform hosts a uniform one, and cluster power rails — when the
# platform exposes them — must decompose power.cpu.
import json
import re
import subprocess

from behave import given, when, then


@given("a built heimdall-daemon binary")
def step_daemon_built(context):
    path = context.bin / "heimdall-daemon"
    assert path.exists(), f"missing binary {path} (make build-tui)"
    context.daemon = str(path)


@when("the daemon collects one snapshot")
def step_collect_once(context):
    out = subprocess.run(
        [context.daemon, "--once", "--json"],
        capture_output=True, text=True, timeout=60, cwd=str(context.root),
    )
    assert out.returncode == 0, f"daemon --once failed: {out.stderr}"
    context.metrics = {}
    for line in out.stdout.splitlines():
        line = line.strip()
        if not line.startswith("{"):
            continue
        m = json.loads(line)
        context.metrics[m.get("metric", "")] = m


@then('the snapshot contains an OK "cpu.topology" metric')
def step_topology_present(context):
    topo = context.metrics.get("cpu.topology")
    assert topo is not None, "cpu.topology missing from snapshot"
    assert topo["status"] == "ok", f"cpu.topology status {topo['status']}"
    context.topology = topo


@then("the topology has exactly one entry per logical core")
def step_topology_length(context):
    cores = context.metrics.get("cpu.cores")
    assert cores is not None and cores.get("cores"), "cpu.cores missing"
    n_topo = len(context.topology.get("cores", []))
    n_cores = len(cores["cores"])
    assert n_topo == n_cores, f"topology entries {n_topo} != cores {n_cores}"


@then("the topology detail describes the core mix")
def step_topology_detail(context):
    detail = context.topology.get("detail", "")
    hybrid = re.fullmatch(r"\d+P( \+ \d+E)?( \+ \d+LP)?", detail)
    uniform = re.fullmatch(r"\d+ cores \(uniform\)", detail)
    assert hybrid or uniform, f"unrecognised topology detail {detail!r}"


@then("every topology entry is a known core-type id")
def step_topology_ids(context):
    for v in context.topology.get("cores", []):
        assert v in (0, 1, 2), f"unknown core-type id {v}"


@then('any "power.cpu.pcluster" and "power.cpu.ecluster" readings are non-negative watts')
def step_cluster_rails_sane(context):
    for name in ("power.cpu.pcluster", "power.cpu.ecluster"):
        m = context.metrics.get(name)
        if m is None:
            continue  # host has no cluster rails — absence is the contract
        assert m["status"] == "ok", f"{name} status {m['status']}"
        assert m["value"] >= 0, f"{name} negative: {m['value']}"
        assert m["unit"] == "watts", f"{name} unit {m['unit']}"


@then('when both cluster rails are OK their sum is close to "power.cpu"')
def step_cluster_sum_matches(context):
    p = context.metrics.get("power.cpu.pcluster")
    e = context.metrics.get("power.cpu.ecluster")
    cpu = context.metrics.get("power.cpu")
    if not (p and e and cpu) or cpu.get("status") != "ok":
        return  # rails absent, or power.cpu came from another source
    if cpu.get("detail", "").find("SMC") == -1:
        return  # power.cpu not SMC-sourced (e.g. IOReport on base dies)
    total = p["value"] + e["value"]
    # same rails read a moment apart — allow drift, catch double-counting
    assert abs(total - cpu["value"]) <= max(2.0, 0.25 * cpu["value"]), (
        f"cluster sum {total} vs power.cpu {cpu['value']}"
    )
