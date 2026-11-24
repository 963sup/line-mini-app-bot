#!/usr/bin/env bash
set -euo pipefail

export PATH="$HOME/.local/bin:$PATH"
if ! gopls help mcp >/dev/null 2>&1; then
	echo 'gopls MCP support is required. Run scripts/setup/install_gopls.sh first.' >&2
	exit 1
fi
codex mcp add gopls -- "$(command -v gopls)" mcp
