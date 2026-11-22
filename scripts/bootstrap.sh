#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source scripts/dev/dev-env.sh
# Setup and explicit dependency refresh must never wait for terminal input.
export MISE_YES=1
export CI=true
mode="${1:-cloud}"
case "$mode" in core|cloud|mcp|all) ;; *) echo "Usage: $0 [core|cloud|mcp|all]" >&2; exit 2;; esac
bash scripts/setup/install-mise.sh
mise trust mise.toml
if [[ "$mode" == cloud ]]; then
  # Keep cache restoration proportional to the runnable application.
  # Analysis tools, Mermaid and MCPs remain explicit opt-in tasks.
  mise run install-runtime
  mise exec -- pnpm install --frozen-lockfile --prefer-offline
  bash scripts/setup/go-deps.sh
  exit 0
fi
mise run install-core
if [[ "$mode" != mcp ]]; then
  mise run install-web
  mise exec -- pnpm install --frozen-lockfile --prefer-offline
  mise run doctor
fi
mise run install-mermaid
if [[ "$mode" == mcp || "$mode" == all ]]; then
  mise run install-mcp
  mise run install-cloud
  mise run graphify-update
fi
