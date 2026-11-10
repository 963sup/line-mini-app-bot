#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source scripts/tooling/dev-env.sh
case "${1:?task required}" in
  dev)
    [[ -f backend/go.mod && -d backend/cmd/server ]] || { echo "Application not created: requires go.mod and cmd/server" >&2; exit 2; }
    export PORT="${PORT:-8080}"
    cd backend
    exec mise exec -- go run ./cmd/server ;;
  deps)
    [[ -f backend/go.mod ]] || { echo "No go.mod; create the application module first" >&2; exit 2; }
    cd backend
    mise exec -- go mod download
    mise exec -- go mod verify ;;
  container-build)
    [[ -f Dockerfile ]] || { echo "No Dockerfile; create application container definition first" >&2; exit 2; }
    exec docker build -t "${LOCAL_IMAGE:-line-mini-app-bot:dev}" . ;;
  *) echo "Unknown task" >&2; exit 2 ;;
esac
