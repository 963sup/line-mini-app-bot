#!/usr/bin/env bash
set -euo pipefail

if (( $# > 1 )) || [[ $# == 1 && $1 != --update ]]; then
	echo "Usage: $0 [--update]" >&2
	exit 2
fi
if ! command -v uv >/dev/null 2>&1; then
	echo 'uv is required to install Graphify.' >&2
	exit 1
fi

if [[ ${1:-} == --update ]] || ! command -v graphify >/dev/null 2>&1; then
	uv tool install --force graphifyy
	hash -r
fi

graphify --version
