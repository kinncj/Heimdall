#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Audit the v2 socket model on a live host: prove the daemon exposes no inbound
# socket and holds only its one outbound stream to the hub, and that the hub is
# the sole listener. Run it on a daemon host (and/or a hub host).
#
# Usage: scripts/verify-sockets.sh
# Exit:  0 = clean, 1 = a violation (a daemon is listening, or has stray conns).
set -euo pipefail

command -v ss >/dev/null || { echo "need: ss (iproute2)"; exit 2; }

# pids by process name, one per line (empty if none running here).
pids() { pgrep -x "$1" 2>/dev/null || true; }

# listening sockets owned by a pid: -l listening, -t tcp / -x unix, -n numeric,
# -H no header, -p process, filtered to this pid's owner tuple.
listen_inet() { ss -H -ltnp 2>/dev/null | grep -F "pid=$1," || true; }
listen_unix() { ss -H -lxnp 2>/dev/null | grep -F "pid=$1," || true; }
estab_inet()  { ss -H -tnp state established 2>/dev/null | grep -F "pid=$1," || true; }

fail=0
note() { printf '  %s\n' "$1"; }
bad()  { printf '  ✗ %s\n' "$1"; fail=1; }
ok()   { printf '  ✓ %s\n' "$1"; }

echo "heimdall socket audit ($(uname -n))"

echo "› heimdall-daemon — must NOT listen; one outbound conn to the hub"
dpids="$(pids heimdall-daemon)"
if [[ -z "$dpids" ]]; then
	note "no daemon running here"
else
	for p in $dpids; do
		li="$(listen_inet "$p")"; lu="$(listen_unix "$p")"
		[[ -n "$li" ]] && bad "pid $p is listening on a TCP port:" && echo "$li" | sed 's/^/      /'
		[[ -n "$lu" ]] && bad "pid $p holds a unix listener:" && echo "$lu" | sed 's/^/      /'
		[[ -z "$li$lu" ]] && ok "pid $p listens on nothing (no inbound surface)"
		est="$(estab_inet "$p")"
		n=$(printf '%s' "$est" | grep -c . || true)
		if [[ "$n" -eq 1 ]]; then
			ok "pid $p has one outbound connection:"; echo "$est" | sed 's/^/      /'
		else
			bad "pid $p has $n established connections (expected 1, to the hub):"
			echo "$est" | sed 's/^/      /'
		fi
	done
fi

echo "› heimdall-hub — the sole network listener"
hpids="$(pids heimdall-hub)"
if [[ -z "$hpids" ]]; then
	note "no hub running here"
else
	for p in $hpids; do
		li="$(listen_inet "$p")"
		[[ -n "$li" ]] && ok "pid $p listens (expected):" && echo "$li" | sed 's/^/      /'
		[[ -z "$li" ]] && bad "pid $p is the hub but holds no listener"
	done
fi

echo "› heimdall-helper — local privileged unix socket only (if running)"
helpids="$(pids heimdall-helper)"
if [[ -z "$helpids" ]]; then
	note "no helper running here"
else
	for p in $helpids; do
		lu="$(listen_unix "$p")"; li="$(listen_inet "$p")"
		[[ -n "$li" ]] && bad "pid $p (helper) listens on a NETWORK port — must be unix-only:" && echo "$li" | sed 's/^/      /'
		[[ -n "$lu" ]] && ok "pid $p listens on a unix socket (expected):" && echo "$lu" | sed 's/^/      /'
	done
fi

echo
if [[ "$fail" -ne 0 ]]; then
	echo "RESULT: VIOLATION — see ✗ above"
	exit 1
fi
echo "RESULT: clean — daemons outbound-only, hub the sole network listener"
