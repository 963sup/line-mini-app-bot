#!/usr/bin/env bash
set -euo pipefail

if (( $# != 2 )) || [[ $1 == -* || $2 == -* ]]; then
	echo "Usage: $0 TOOL_PACKAGE DEPENDENCY" >&2
	exit 2
fi
if ! command -v uv >/dev/null 2>&1; then
	echo 'uv is required to inject a tool dependency.' >&2
	exit 1
fi

tool=$1
dependency=$2
cache=$(mktemp -d)
trap 'rm -rf -- "$cache"' EXIT

if uv tool --help | grep -Eq '^[[:space:]]+inject([[:space:]]|$)'; then
	UV_CACHE_DIR="$cache" uv tool inject "$tool" "$dependency"
else
	UV_CACHE_DIR="$cache" uv tool install --force --link-mode=copy --with "$dependency" "$tool"
fi
