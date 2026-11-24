#!/usr/bin/env bash

set -euo pipefail

echo 'export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
hash -r
go install golang.org/x/tools/gopls@latest
codex mcp add gopls -- gopls mcp

codex mcp add openaiDeveloperDocs --url https://developers.openai.com/mcp
codex mcp add context7 --url https://mcp.context7.com/mcp
codex mcp add shadcn -- npx -y shadcn@latest mcp
codex mcp add next-devtools -- npx -y next-devtools-mcp@latest
codex mcp add github --url https://api.githubcopilot.com/mcp/ --bearer-token-env-var GITHUB_MCP_TOKEN
codex mcp add repomix -- npx -y repomix --mcp

codex mcp list
