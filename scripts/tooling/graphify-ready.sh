#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.."

if ! command -v graphify >/dev/null 2>&1; then
	echo 'Graphify CLI is required. Run scripts/setup/install-graphify.sh first.' >&2
  exit 1
fi

if [[ -s graphify-out/graph.json ]]; then
  graphify update .
else
  graphify extract . --code-only
fi

if [[ ! -s graphify-out/graph.json ]]; then
  echo 'Graphify did not produce graphify-out/graph.json.' >&2
  exit 1
fi

graphify cluster-only .
