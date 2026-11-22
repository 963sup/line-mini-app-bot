#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source scripts/dev/dev-env.sh
[[ -f backend/go.mod ]] || { echo "No go.mod; create the application module first" >&2; exit 2; }
cd backend
mise exec -- sh -ec 'go mod download && go mod verify'
