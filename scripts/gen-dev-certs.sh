#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Generate a self-signed TLS cert + key for local Heimdall development.
#
# The one cert is used two ways: the hub presents it as its server cert
# (--tls-cert/--tls-key) and daemons/dashboards trust it as their CA bundle
# (--tls-ca). So it must carry SANs for the names clients dial — localhost and
# 127.0.0.1. DEV ONLY: do not use these certs anywhere real.
#
# Usage: gen-dev-certs.sh [DIR]   (DIR defaults to ./certs)
set -euo pipefail

dir="${1:-certs}"
mkdir -p "$dir"

openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout "$dir/hub.key" -out "$dir/hub.crt" -days 3650 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" >/dev/null 2>&1

chmod 600 "$dir/hub.key"
echo "wrote $dir/hub.crt and $dir/hub.key (self-signed, dev only)"
