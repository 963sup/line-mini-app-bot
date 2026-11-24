#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
path_line='export PATH="$HOME/.local/bin:$HOME/.local/go/bin:$HOME/go/bin:$PATH"'
touch "$HOME/.bashrc"
grep -Fqx "$path_line" "$HOME/.bashrc" || printf '%s\n' "$path_line" >> "$HOME/.bashrc"
export PATH="$HOME/.local/bin:$HOME/.local/go/bin:$HOME/go/bin:$PATH"
hash -r

"$root/scripts/setup/install-gopls.sh"
codex mcp add gopls -- gopls mcp

codex mcp add openaiDeveloperDocs --url https://developers.openai.com/mcp
codex mcp add context7 --url https://mcp.context7.com/mcp
codex mcp add shadcn -- npx -y shadcn@latest mcp
codex mcp add next-devtools -- npx -y next-devtools-mcp@latest
codex mcp add github --url https://api.githubcopilot.com/mcp/ --bearer-token-env-var GITHUB_MCP_TOKEN

codex mcp list
