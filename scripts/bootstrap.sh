#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
# Setup and explicit dependency refresh must never wait for terminal input.
export MISE_YES=1
export CI=true
mode="${1:-cloud}"
case "$mode" in core|cloud) ;; *) echo "Usage: $0 [core|cloud]" >&2; exit 2;; esac
source scripts/dev/dev-env.sh
# Bootstrap installs only the tool groups selected below.
export MISE_TASK_RUN_AUTO_INSTALL=false
export MISE_EXEC_AUTO_INSTALL=false
bash scripts/setup/install-mise.sh
mise trust mise.toml
# Install the toolchain for the selected mode before restoring dependencies.
if [[ "$mode" == cloud ]]; then
  mise run install-runtime
else
  mise run install-core
  mise run install-web
fi

mise exec -- pnpm install --frozen-lockfile --prefer-offline
if [[ "$mode" == cloud ]]; then
  exec bash scripts/setup/go-deps.sh
fi

mise run doctor
mise run install-mermaid
