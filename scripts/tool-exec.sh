#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/cloud-env.sh"
cd "$(dirname "${BASH_SOURCE[0]}")/.."
exec mise exec -- "$@"
