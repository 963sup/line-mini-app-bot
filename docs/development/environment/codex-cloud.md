# Codex Cloud 環境

建立或更新此專案的 Codex Cloud 環境時，以儲存庫根目錄為 checkout，將以下命令設為 **Install script**。

```bash
scripts/bootstrap.sh
```

預設 `cloud` 模式依版本與鎖檔準備 Go、Node、pnpm 和 API 產生器，安裝 workspace 依賴並下載、驗證 Go modules。初始化會重用已存在的工具與套件快取。Go 分析、Mermaid 和 MCP 按任務準備。Cloud 模式只準備環境。依據：[初始化腳本](../../../scripts/bootstrap.sh) · [工具任務](../../../mise.toml)。

Linux x64／arm64 下載資訊由 [mise.lock](../../../mise.lock) 鎖定，pnpm CLI 的相依鎖在 `.mise/locks/`，應用依賴以 [pnpm-lock.yaml](../../../pnpm-lock.yaml) 和 [backend/go.sum](../../../backend/go.sum) 驗證。選用工具有自己的版本來源。

## 已初始化的工作區

開始任務時，先用 `git rev-parse --show-toplevel` 確認 checkout，並在該根目錄執行 `bash scripts/tooling/tool-exec.sh mise ls --installed` 查看可用版本。沿[模組導覽](../../../README.md#巢狀目錄導覽)讀目標路徑的工作指引；跨模組任務依 [Codex 閱讀順序](../tooling/agents/codex.md#從根目錄處理跨模組任務)核對程式與規格。若缺少單一工具，依[根目錄工具指引](../../../AGENTS.md#按需安裝工具)準備。

## 啟動與更新

**Start skill** 應從 checkout 根目錄透過專案 wrapper 執行命令。需使用應用時，依[啟動表](README.md#啟動應用)選服務，核對目標路由及日誌。發布狀態保存已準備的檔案與快取，執行中的程序需於工作階段啟動。

Codex CLI 與 Cloud 任務提交方式見 [Codex 操作](../tooling/agents/codex.md#開啟專案與提交-cloud-任務)。

工具設定或依賴鎖檔變更時，重新執行初始化並發布更新的環境。從新聊天驗收已發布的狀態。若需比較速度，分開記錄空快取準備與重用快取時的耗時。

## 新環境驗收

1. 在新 checkout 確認安裝工具與工作區依賴。若還要執行完整專案 `pnpm check` 或 Go 格式及 lint，先透過 wrapper 執行 `mise run install-core`。
2. 執行 `bash scripts/tooling/tool-exec.sh mise run check` 與 `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check`，並依任務驗收實際入口。
3. 在新聊天依 [MCP 主機載入檢查](../tooling/agents/README.md#project-loading-check)。
4. 若任務需要 Cloud Run 或 Firestore Toolbox，依主機方式提供 ADC。使用憑證檔時，`GOOGLE_APPLICATION_CREDENTIALS` 須指向任務執行時可讀的檔案。以目標資源的唯讀操作確認登入身分及 IAM。參見 [Google 工具指引](../tooling/shared/gcloud.md)。

記錄所驗證的快照、命令、結果與尚缺的環境證據。[驗收索引](../verification/README.md) · [環境排查](troubleshooting.md)
