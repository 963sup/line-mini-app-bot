# 環境排查

準備或啟動失敗時，記錄 checkout、命令、實際錯誤與發生階段，再按症狀選一個檢查點。

| 症狀 | 下一個檢查 |
| --- | --- |
| 找不到 CLI 或版本不符 | 從根目錄使用 `bash scripts/tooling/tool-exec.sh mise ls --installed`，對照 [mise 工具入口](../tooling/shared/mise.md)與鎖檔 |
| 安裝失敗 | 確認[系統前置條件](README.md)、原始鎖檔與套件網路可用，記錄是哪一項下載或驗證失敗 |
| Next.js 頁面可開，但 API 失敗 | 依[啟動表](README.md#啟動應用)確認 Go 服務、8080 埠與實際 API 回應，對照兩邊日誌 |
| 設定存在，但聊天找不到 MCP 工具 | 依[主機載入檢查](../tooling/agents/README.md#project-loading-check)分別核對信任、載入、transport 與原生工具 |
| 已看得到工具，但呼叫資源時被拒絕 | 依[工具指引](../tooling/shared/gcloud.md)確認實際身分、專案／資料庫、資源權限與錯誤訊息 |

將確認的失敗階段及能重現的命令寫入交付。若需檢查程式根因，轉入[系統化除錯](../workflow/debugging.md#diagnostic-loop)。

## 設定與 MCP 診斷

依 [MCP 診斷](../tooling/agents/README.md#mcp-診斷)核對設定、測試服務連線並確認主機載入。

Codex 設定查 [OpenAI Docs MCP](../tooling/agents/openai-docs.md)，版本專屬行為查 [Context7](../tooling/agents/context7.md)。維持主機提供的 `HOME`、`CODEX_HOME` 與憑證設定，記錄錯誤時遮蔽認證標頭和機密值。

[環境索引](README.md)
