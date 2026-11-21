#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source scripts/dev/dev-env.sh
[[ -f backend/go.mod && -f backend/main.go ]] || { echo "Application not created: requires go.mod and main.go" >&2; exit 2; }
export DEV_ERROR_DETAILS=1
export PORT="${PORT:-8080}"
cd backend
exec mise exec -- go run .
