# 環境排查

準備或啟動失敗時，記錄 checkout、命令、實際錯誤與發生階段，再按症狀選一個檢查點。

| 症狀                                         | 下一個檢查                                                                                                                                                                                  |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Codex CLI 跳過專案指引或提早結束             | 在出問題的啟動目錄執行 `codex debug prompt-input`，於本機核對模型輸入是否包含預期的 AGENTS.md。缺少時查啟動目錄與指引路徑；完整載入時，從原任務紀錄找出第一次偏離目標的工具選擇或停止決定。 |
| 找不到 CLI 或版本不符                        | 從根目錄使用 `bash scripts/tooling/tool-exec.sh mise ls --installed`，對照 [mise 工具入口](../tooling/shared/mise.md)與鎖檔                                                                 |
| Cloud 初始化顯示 `Cannot install mise for …` | 用 `uname -s -m` 核對執行 Install script 的主機；[Cloud 初始化](codex-cloud.md)使用 Linux x86_64 或 Linux aarch64 的發行檔                                                                  |
| 安裝失敗                                     | 確認[系統前置條件](../tooling/shared/toolchain-setup.md)、原始鎖檔與套件網路可用，記錄是哪一項下載或驗證失敗                                                                                |
| Next.js 頁面可開，但 API 失敗                | 依[啟動表](codex-cloud.md#本機服務啟動)確認 Go 服務、8080 埠與實際 API 回應，對照兩邊日誌                                                                                                   |
| 設定存在，但聊天找不到 MCP 工具              | 依[主機載入檢查](#設定與-mcp-診斷)分別核對信任、載入、transport 與原生工具                                                                                                                  |
| 已看得到工具，但呼叫資源時被拒絕             | 依[工具指引](../tooling/shared/gcloud.md)確認實際身分、專案／資料庫、資源權限與錯誤訊息                                                                                                     |

將確認的失敗階段及能重現的命令寫入交付。若需檢查程式根因，轉入[系統化除錯](../workflow/debugging.md#diagnostic-loop)。

## 設定與 MCP 診斷

依序區分設定、服務連線、主機載入與功能驗收，不以其中一項成功代替其他項目：

1. 核對 [.codex/config.toml](../../../.codex/config.toml) 的服務名稱、transport 與 allowlist。Codex CLI 可透過 `bash scripts/tooling/tool-exec.sh python3 scripts/mcp/check-host-config.py` 比對有效註冊；此結果不代表桌面聊天已載入工具。
2. 對有問題的服務執行 `bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server <服務名稱> --list-tools`，檢查 initialize、tools/list 及工具 schema。只有需要全服務診斷且具備網路與必要憑證時，才執行 `mise run doctor-mcp`。
3. 在目前聊天檢查實際曝光的工具。若缺少，核對主機選用的專案、信任與載入設定，再使用主機支援的重新載入方式。專案 CLI 註冊與主機 connector 可能不同，以實際 schema 為準。
4. 使用任務所需的代表性唯讀操作驗證目標資源。記錄 host、server、失敗階段與原始錯誤；條件未變時沿用已診斷結果，不反覆重試逾時。

實作來源：[CLI 註冊比對](../../../scripts/mcp/check-host-config.py)與[服務診斷](../../../scripts/mcp/mcp-doctor.py)。

Codex 設定查 [OpenAI Docs MCP](../tooling/agents/openai-docs.md)，版本專屬行為查 [Context7](../tooling/agents/context7.md)。維持主機提供的 `HOME`、`CODEX_HOME` 與憑證設定，記錄錯誤時遮蔽認證標頭和機密值。

[環境索引](./)
