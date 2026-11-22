# mise ownership

## Go development loop

Go 工具準備與開發操作見[後端指南](../backend/go-development.md#go-development-loop)。

## 共用版本與任務

Use `bash scripts/tooling/tool-exec.sh mise run <task>` from the repository root.
Tool versions, non-secret paths and tasks are defined in `mise.toml`.
The wrapper loads bootstrap-compatible environment before invoking mise.
Do not change platform HOME/CODEX_HOME or copy authentication files.

- `bootstrap`: core installation and Go doctor.
- `install-mcp`: optional local MCP tools and generated Graphify skills.
- `install-shadcn`：只安裝固定 shadcn CLI 與 Node runtime，不生成元件或技能；CLI 與其 MCP 共用 mise 的固定工具版本，UI/runtime 依賴由前端 package 與 lockfile 管理，見 [shadcn UI 與 MCP](../frontend/shadcn.md)。
- `install-next-devtools`：安裝固定 Next.js DevTools MCP connector 與 Node；runtime 查詢另需本機 Next.js 開發伺服器，見 [Next DevTools](../frontend/nextjs-devtools.md)。
- `graphify-update` / `repomix`: refresh analysis artifacts.
- `check-config` / `doctor`: configuration and Go environment verification.
- `doctor-mcp`: configured service transports; needs network and GitHub token.
- `doctor-openai`: isolated documentation transport diagnosis.
- `firestore-emulator`：按需準備固定版本的 Java 25，啟動本機 Firestore emulator。見[操作與測試](../testing/firestore-emulator.md#firestore-emulator)。
- `install-cloud`: pinned gcloud CLI and local gcloud MCP installation.
- `doctor-gcloud` / `doctor-google-knowledge`: separate Google MCP transport checks; see [Google setup](../agents/google.md).
- `check` / `build`: application checks; require an existing Go module.

Use `mise tasks` to discover current tasks. Bootstrap needs system Python 3.11+,
curl, archive/checksum utilities and network access before mise is installed.
Workspace tool paths target Codex Cloud; arbitrary machines need deliberate path
configuration. Existing cache success is not a clean rebuild test.

## 鎖檔與安裝結果

已有 `mise.lock` 時，mise 安裝單一工具也可能解析整份工具設定；任何失敗會中止後續批次步驟，不能只憑套件下載完成宣稱整個安裝任務成功。診斷時分別驗證任務 exit code、CLI 執行與 MCP handshake；MCP 通過不代表聊天已載入或 Next.js runtime 已啟動。
