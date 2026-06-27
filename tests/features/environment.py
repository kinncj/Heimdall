# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# behave environment: build the binaries once, then start and tear down real
# heimdall processes per scenario. The acceptance suite drives the shipped
# binaries end to end — no mocks.
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[2]
BIN = ROOT / "bin"


def before_all(context):
    subprocess.run(
        ["make", "build-tui"],
        cwd=str(ROOT),
        check=True,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    context.root = ROOT
    context.bin = BIN


def before_scenario(context, scenario):
    context.procs = []


def after_scenario(context, scenario):
    for proc in getattr(context, "procs", []):
        try:
            proc.terminate()
            proc.wait(timeout=5)
        except Exception:
            try:
                proc.kill()
            except Exception:
                pass
