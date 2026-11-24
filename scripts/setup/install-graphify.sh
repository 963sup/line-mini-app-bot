#!/usr/bin/env bash
set -euo pipefail

if (( $# > 1 )) || [[ $# == 1 && $1 != --update ]]; then
	echo "Usage: $0 [--update]" >&2
	exit 2
fi
"$(dirname -- "${BASH_SOURCE[0]}")/install-graphify-anyio.sh" "${@}"

graphify --version
