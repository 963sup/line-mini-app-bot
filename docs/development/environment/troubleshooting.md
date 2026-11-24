# 環境排查

準備或啟動失敗時，記錄 checkout、命令、實際錯誤與發生階段，再按症狀選一個檢查點。

| 症狀                                         | 下一個檢查                                                                                                                                                                                  |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Codex CLI 跳過專案指引或提早結束             | 在出問題的啟動目錄執行 `codex debug prompt-input`，於本機核對模型輸入是否包含預期的 AGENTS.md。缺少時查啟動目錄與指引路徑；完整載入時，從原任務紀錄找出第一次偏離目標的工具選擇或停止決定。 |
| 找不到 CLI 或版本不符                        | 從根目錄使用 `bash scripts/tooling/tool-exec.sh mise ls --installed`，對照 [mise 工具入口](../tooling/shared/mise.md)與鎖檔                                                                 |
| Cloud 初始化顯示 `Cannot install mise for …` | 用 `uname -s -m` 核對執行 Install script 的主機；[Cloud 初始化](codex-cloud.md)使用 Linux x86_64 或 Linux aarch64 的發行檔                                                                  |
| 安裝失敗                                     | 確認[系統前置條件](../tooling/shared/mise.md#共用版本與任務)、原始鎖檔與套件網路可用，記錄是哪一項下載或驗證失敗                                                                            |
| Next.js 頁面可開，但 API 失敗                | 依[啟動表](codex-cloud.md#本機服務啟動)確認 Go 服務、8080 埠與實際 API 回應，對照兩邊日誌                                                                                                   |
| 設定存在，但聊天找不到 MCP 工具              | 依[主機載入檢查](#設定與-mcp-診斷)分別核對信任、載入、transport 與原生工具                                                                                                                  |
| 已看得到工具，但呼叫資源時被拒絕             | 依[工具指引](../tooling/shared/gcloud.md)確認實際身分、專案／資料庫、資源權限與錯誤訊息                                                                                                     |

將確認的失敗階段及能重現的命令寫入交付。若需檢查程式根因，轉入[除錯與根因驗證](../workflow/implementation.md#diagnostic-loop)。

## 設定與 MCP 診斷

依序區分設定、服務連線、主機載入與功能驗收，不以其中一項成功代替其他項目：

1. 先查目前聊天是否曝光所需工具。若缺少，核對 [.codex/config.toml](../../../.codex/config.toml) 的服務名稱、transport 與 allowlist。從受信任的儲存庫根目錄使用 `codex mcp list` 核對註冊，以 `codex mcp get <服務名稱>` 檢視設定；此結果不代表桌面聊天已載入工具。
2. 在互動式 Codex 使用 `/mcp` 查看服務狀態與工具清單。
3. 專案已定義服務時，先確認主機選用及信任此 checkout，並從根目錄重新開啟 Codex。若 CLI 已列出服務但聊天仍缺少工具，檢查 transport、工具篩選與主機載入；主機 connector 的工具以實際 schema 為準。
4. 確認專案設定已載入、必要服務仍缺少註冊時，HTTP 服務使用 `codex mcp add <name> --url <url>`；本機服務先以專案工具鏈準備指定版本，再用 `codex mcp add <name> -- <command> [args...]`。這會寫入使用者層設定，須保留專案指定的執行目錄、環境與核准設定，並重新開啟聊天核對工具。
5. 使用任務所需的代表性唯讀操作驗證目標資源。記錄 host、server、失敗階段與原始錯誤；條件未變時沿用已診斷結果，不反覆重試逾時。

Codex 設定查 [Codex 文件查證](../tooling/agents/codex.md#codex-與-openai-文件查證)，版本專屬行為查 [Context7](../tooling/agents/context7.md)。維持主機提供的 `HOME`、`CODEX_HOME` 與憑證設定，記錄錯誤時遮蔽認證標頭和機密值。

[環境索引](./)
