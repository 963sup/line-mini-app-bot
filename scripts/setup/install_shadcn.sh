#!/usr/bin/env bash
set -euo pipefail

prefix="$HOME/.local/share/codex-mcp/shadcn"
if (( $# > 1 )) || [[ $# == 1 && $1 != --update && $1 != --register-only ]]; then
  echo "Usage: $0 [--update | --register-only]" >&2
  exit 2
fi
if [[ ${1:-} == --register-only ]]; then
  if [[ ! -x "$prefix/bin/shadcn" ]]; then
    echo 'shadcn is required. Run scripts/setup/install_shadcn.sh first.' >&2
    exit 1
  fi
elif [[ ${1:-} == --update || ! -x "$prefix/bin/shadcn" ]]; then
  cache=$(mktemp -d)
  trap 'rm -rf -- "$cache"' EXIT
  npm install --global --prefix "$prefix" --cache "$cache" --no-audit --no-fund 'shadcn@latest'
fi
codex mcp add shadcn -- "$prefix/bin/shadcn" mcp
