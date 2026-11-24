#!/usr/bin/env bash
set -euo pipefail

codex mcp add github --url https://api.githubcopilot.com/mcp/ --bearer-token-env-var GITHUB_MCP_TOKEN
