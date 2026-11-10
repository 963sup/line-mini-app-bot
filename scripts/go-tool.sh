#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec bash "$root/scripts/tool-exec.sh" bash -c 'cd "$1/backend"; shift; exec go "$@"' -- "$root" "$@"
