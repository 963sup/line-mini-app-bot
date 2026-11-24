#!/usr/bin/env bash
set -euo pipefail

prefix="$HOME/.local/share/codex-mcp/repomix"
if (( $# > 1 )) || [[ $# == 1 && $1 != --update ]]; then
  echo "Usage: $0 [--update]" >&2
  exit 2
fi
if [[ ${1:-} == --update || ! -x "$prefix/bin/repomix" ]]; then
  cache=$(mktemp -d)
  trap 'rm -rf -- "$cache"' EXIT
  npm install --global --prefix "$prefix" --cache "$cache" --no-audit --no-fund 'repomix@latest'
fi
codex mcp add repomix -- "$prefix/bin/repomix" --mcp
