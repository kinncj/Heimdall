# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# behave environment: build the binaries once, then start and tear down real
# heimdall processes per scenario. The acceptance suite drives the shipped
# binaries end to end — no mocks.
import os
import pathlib
import shutil
import subprocess
import tempfile

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
    # Isolate each scenario's config in a throwaway dir. The daemon persists
    # settings (control-listen, log-source, token, ...) to a shared user config
    # path whenever a flag is passed; without isolation one scenario's saved
    # config leaks into the next — e.g. the control-plane scenario's
    # control-listen would make the enrollment daemons collide on a port and
    # exit. This also keeps the suite off the developer's real ~/.config.
    context.config_dir = tempfile.mkdtemp(prefix="heimdall-cfg-")
    os.environ["HEIMDALL_CONFIG_DIR"] = context.config_dir


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
    os.environ.pop("HEIMDALL_CONFIG_DIR", None)
    shutil.rmtree(getattr(context, "config_dir", ""), ignore_errors=True)
