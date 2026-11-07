#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/cloud-env.sh"
cd /workspace/line-mini-app-bot
exec mise exec -- "$@"
