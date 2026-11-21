#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source scripts/dev/dev-env.sh
[[ -f Dockerfile ]] || { echo "No Dockerfile; create application container definition first" >&2; exit 2; }
build_options=()
if [[ -n ${BUILD_CA_FILE:-} ]]; then
  [[ -r "$BUILD_CA_FILE" ]] || { echo "BUILD_CA_FILE must be readable" >&2; exit 2; }
  build_options+=(--secret "id=build-ca,src=$BUILD_CA_FILE")
fi
docker build "${build_options[@]}" --build-arg VCS_REF="$(git rev-parse HEAD)" -t "${LOCAL_IMAGE:-line-mini-app-bot:dev}" .
