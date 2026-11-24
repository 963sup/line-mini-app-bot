#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.."
cache=$(mktemp -d)
trap 'rm -rf -- "$cache"' EXIT
npm_config_cache="$cache" npx --yes skills add yamadashy/repomix --skill repomix-explorer --agent codex --yes
