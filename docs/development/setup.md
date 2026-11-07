# Environment setup

From the repository root:

```bash
bash scripts/bootstrap.sh core
bash scripts/tool-exec.sh mise run check-config
bash scripts/tool-exec.sh mise run doctor
# Optional tools, separate from core Go development:
bash scripts/bootstrap.sh mcp
```

See [mise prerequisites](../tooling/mise.md). GitHub MCP requires the runtime
credential named in `.codex/config.toml`; never put its value in tracked files.
MCP installation and transport success do not register tools in the chat host.
Application commands require go.mod and application source; do not create them
merely to make environment checks pass.
