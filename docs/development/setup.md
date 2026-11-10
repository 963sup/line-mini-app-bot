# Environment setup

Current priority is the execution/deployment foundation, not attendance behavior.
Follow [the phased foundation plan](foundation.md) for prerequisites, current
evidence and completion criteria before implementing [attendance](../features/attendance.md).

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

## Monorepo 工作目錄

Go module 位於 `backend/`；直接執行 Go 命令請先 `cd backend`。mise 的 Go 任務已指定工作目錄。前端與契約命令由儲存庫根目錄執行；流程見 [Monorepo 決策](../architecture/decisions/0004-monorepo.md)。
