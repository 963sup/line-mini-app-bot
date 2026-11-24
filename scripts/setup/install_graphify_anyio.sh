#!/usr/bin/env bash
set -euo pipefail

if (( $# > 1 )) || [[ $# == 1 && $1 != --update ]]; then
	echo "Usage: $0 [--update]" >&2
	exit 2
fi
has_anyio=false
if command -v graphify >/dev/null 2>&1; then
	graphify_python=$(sed -n '1s/^#!//p' "$(command -v graphify)")
	if [[ -x $graphify_python ]] && "$graphify_python" -c 'import anyio' >/dev/null 2>&1; then
		has_anyio=true
	fi
fi

if [[ ${1:-} == --update ]] || [[ $has_anyio != true ]]; then
	"$(dirname -- "${BASH_SOURCE[0]}")/inject_uv_tool_dependency.sh" graphifyy anyio
	hash -r
fi

graphify_python=$(sed -n '1s/^#!//p' "$(command -v graphify)")
"$graphify_python" -c 'import anyio'
