#!/usr/bin/env bash
set -euo pipefail

export PATH="$HOME/.local/bin:$PATH"
if ! gopls help mcp >/dev/null 2>&1; then
	echo 'gopls MCP support is required. Run scripts/setup/install-gopls.sh first.' >&2
	exit 1
fi
codex mcp add gopls -- gopls mcp

codex mcp add openaiDeveloperDocs --url https://developers.openai.com/mcp
codex mcp add context7 --url https://mcp.context7.com/mcp
codex mcp add shadcn -- npx -y shadcn@latest mcp
codex mcp add next-devtools -- npx -y next-devtools-mcp@latest
codex mcp add github --url https://api.githubcopilot.com/mcp/ --bearer-token-env-var GITHUB_MCP_TOKEN

codex mcp list
