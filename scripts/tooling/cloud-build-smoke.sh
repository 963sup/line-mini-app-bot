#!/usr/bin/env bash
# Exercise the actual Cloud Build step image without credentials or cloud writes.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
image=gcr.io/google.com/cloudsdktool/google-cloud-cli:592.0.0-slim
python=/usr/lib/google-cloud-sdk/platform/bundledpythonunix/bin/python3
docker run --rm --network=none --read-only \
  --mount "type=bind,src=$PWD/scripts/tooling,dst=/checks,readonly" \
  --entrypoint "$python" "$image" -B /checks/cloud-build-latest.py --help
docker run --rm --network=none --read-only \
  --mount "type=bind,src=$PWD/scripts/tooling,dst=/checks,readonly" \
  --entrypoint "$python" "$image" -B -m unittest discover \
  -s /checks -p test_cloud_build_latest.py
echo 'PASS: actual Cloud CLI image interpreter, guard startup and cancellation tests'
