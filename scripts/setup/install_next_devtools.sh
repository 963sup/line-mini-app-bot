#!/usr/bin/env bash
set -euo pipefail

prefix="$HOME/.local/share/codex-mcp/next-devtools-mcp"
if (( $# > 1 )) || [[ $# == 1 && $1 != --update && $1 != --register-only ]]; then
  echo "Usage: $0 [--update | --register-only]" >&2
  exit 2
fi
if [[ ${1:-} == --register-only ]]; then
  if [[ ! -x "$prefix/bin/next-devtools-mcp" ]]; then
    echo 'next-devtools-mcp is required. Run scripts/setup/install_next_devtools.sh first.' >&2
    exit 1
  fi
elif [[ ${1:-} == --update || ! -x "$prefix/bin/next-devtools-mcp" ]]; then
  cache=$(mktemp -d)
  trap 'rm -rf -- "$cache"' EXIT
  npm install --global --prefix "$prefix" --cache "$cache" --no-audit --no-fund 'next-devtools-mcp@latest'
fi
codex mcp add next-devtools -- "$prefix/bin/next-devtools-mcp"
