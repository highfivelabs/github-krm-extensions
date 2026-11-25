#!/usr/bin/env bash
# Copyright 2026 High Five Labs
# SPDX-License-Identifier: Apache-2.0
# Download the vendor's OpenAPI specs at the commits vendor.yaml pins.
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p specs
[ -s specs/github.json ] || curl -fsSL -o specs/github.json "https://raw.githubusercontent.com/github/rest-api-description/2a33d9cc5d1dc6ca603b908e0796454af88c1999/descriptions/api.github.com/api.github.com.json"
