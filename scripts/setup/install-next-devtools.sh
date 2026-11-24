#!/usr/bin/env bash
set -euo pipefail

prefix="$HOME/.local/share/codex-mcp/next-devtools-mcp"
if (( $# > 1 )) || [[ $# == 1 && $1 != --update ]]; then
  echo "Usage: $0 [--update]" >&2
  exit 2
fi
if [[ ${1:-} == --update || ! -x "$prefix/bin/next-devtools-mcp" ]]; then
  cache=$(mktemp -d)
  trap 'rm -rf -- "$cache"' EXIT
  npm install --global --prefix "$prefix" --cache "$cache" --no-audit --no-fund 'next-devtools-mcp@latest'
fi
codex mcp add next-devtools -- "$prefix/bin/next-devtools-mcp"
