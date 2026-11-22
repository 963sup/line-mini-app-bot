# 開發環境

從儲存庫根目錄執行命令。Linux x64／arm64 需要 Bash、curl、tar、sha256sum 與 Python 3.11 以上。

| 場景 | 操作入口 |
| --- | --- |
| 準備或更新發布的 Codex Cloud 環境 | [Cloud 安裝、啟動與驗收](codex-cloud.md) |
| 本機開發 | `bash scripts/bootstrap.sh core`，再按需求[啟動應用](#啟動應用) |
| 本機需要 MCP 與雲端工具 | [初始化模式](setup.md#一般開發與-mcp-安裝) |
| 核對設定、啟動及分階段驗收 | [環境驗證](verification.md) |
| 工具找不到、依賴或認證失敗 | [環境排查](troubleshooting.md) |

## 啟動應用

確認鎖定依賴已準備好，並依任務所需提供應用設定。分開執行下列命令，方便判斷問題來自 Next.js 或 Go。需要 API 的頁面，讓 Go 服務在預設 8080 埠提供請求。

| 需要驗證 | 啟動命令 | 就緒依據 |
| --- | --- | --- |
| Next.js 頁面 | `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp dev` | 在 3000 埠開啟本次修改的頁面 |
| Go HTTP 或瀏覽器呼叫的 API | `bash scripts/tooling/tool-exec.sh mise run dev` | 日誌顯示服務已啟動，實際請求所需路由並核對回應 |
| Go 提供靜態頁面的整合 | 先用 `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp build` 生成靜態資源，再啟動 Go | 透過 Go 服務開啟目標頁面，核對資源與 API |

閱讀各程序的終端輸出。以前景程序執行時，用 Ctrl+C 結束。啟動與轉送依據：[Go 入口](../../../scripts/dev/run.sh) · [前端任務](../../../frontend/miniapp/package.json) · [開發環境 API 轉送](../../../frontend/miniapp/next.config.ts)。

[開發手冊](../README.md)
