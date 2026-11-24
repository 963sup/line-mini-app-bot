# mise ownership

## 環境初始化與工作目錄

從儲存庫根目錄執行 `bash scripts/bootstrap.sh`（預設 Cloud 模式）或 `bash scripts/bootstrap.sh core`（本機開發），再以 `bash scripts/tooling/tool-exec.sh mise run check-config` 檢查設定。兩種初始化只準備該模式的工具與 workspace 依賴；套件下載及遠端 MCP 需要網路。按需使用的 MCP 另以本頁任務安裝，並依[MCP 診斷](../../environment/troubleshooting.md#設定與-mcp-診斷)確認聊天實際載入。`GITHUB_MCP_TOKEN` 由主機於執行期安全注入。[Codex Cloud 手冊](../../environment/codex-cloud.md)維護 Install script、更新與新聊天驗收。

前端、契約與 `pnpm` 從儲存庫根目錄執行；直接執行 `go` 前進入 `backend/`。Go race 檢查需要 C 編譯器，容器建置需要 Docker daemon。VS Code 的 Go／gopls 工作使用專案 wrapper，開啟 Go 檔前執行 `Go: setup`。[驗收矩陣](../testing/commands.md)決定本次應跑哪些檢查。

## Go development loop

Go 工具準備與開發操作見[後端指南](../backend/go-development.md#go-development-loop)。

## 共用版本與任務

Use `bash scripts/tooling/tool-exec.sh mise run <task>` from the repository root.
Tool versions, non-secret paths and tasks are defined in `mise.toml`.
The wrapper loads bootstrap-compatible environment before invoking mise.
Do not change platform HOME/CODEX_HOME or copy authentication files.

- `bootstrap`: core installation and Go doctor.
- `install-shadcn`：只安裝固定 shadcn CLI 與 Node runtime，不生成元件或技能；CLI 與其 MCP 共用 mise 的固定工具版本，UI/runtime 依賴由前端 package 與 lockfile 管理，見 [shadcn UI 與 MCP](../frontend/shadcn.md)。
- `graphify-update` / `repomix`: refresh analysis artifacts.
- `check-config` / `doctor`: configuration and Go environment verification.
- `firestore-emulator`：按需準備固定版本的 Java 25，啟動本機 Firestore emulator。見[操作與測試](../testing/firestore-emulator.md#firestore-emulator)。
- `install-cloud`: 依 mise 任務安裝 Node、gcloud CLI、Firebase CLI 與本機 gcloud MCP。
- `check` / `build`: application checks; require an existing Go module.

Use `mise tasks` to discover current tasks. Bootstrap needs system Python 3.11+,
curl, archive/checksum utilities and network access before mise is installed.
Workspace tool paths target Codex Cloud; arbitrary machines need deliberate path
configuration. Existing cache success is not a clean rebuild test.

## 鎖檔與安裝結果

已有 `mise.lock` 時，mise 安裝單一工具也可能解析整份工具設定；任何失敗會中止後續批次步驟，不能只憑套件下載完成宣稱整個安裝任務成功。診斷時分別驗證任務 exit code、CLI 執行與 MCP handshake；MCP 通過不代表聊天已載入或 Next.js runtime 已啟動。
