# Codex 專案配置

## 目錄導覽與維護

config.toml 定義角色與 MCP 連線，agents/ 定義角色指引，hooks.json 保存 hooks；skills/ 為安裝產物。

連線設定只在 config.toml 維護；工具版本、任務與非機密環境以 ../mise.toml 為準。不要手改上游技能或寫入憑證。

從根目錄執行 bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py；MCP transport 與聊天主機曝光需分別驗證。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/foundation/architecture/business-operation.md) 與 [符合性流程](../docs/agents/workflow.md#business-operation-compliance)。

現有角色依 Business Operation 契約分工；角色開始相關任務時須讀取模型及符合性流程，文件路徑不是自動 include。

## 角色配置與派工

角色檔位於 `agents/`，由 `config.toml` 註冊；完整職責、派工順序、深度與單一寫入者規則見[代理工作流程](../docs/agents/workflow.md#delegation-depth-and-role-selection)。`graphify`、`repomix` 提供探索證據，`context7` 查證適用版本文件，`golang`、`typescript`、`nextjs`、`shadcn` 接受明確範圍的實作指派，`documentation` 維護 Markdown。工具操作依[MCP 指南](../docs/tooling/mcp/README.md)，不在角色檔重複連線設定。

支援自訂代理的 Codex 可選用已註冊角色；其他主機依工作流程採用明示的 fallback。配置存在不代表目前聊天主機已載入角色或 MCP。角色沿用父代理權限，不增加推送、發布或部署授權。

## Mermaid

`mermaid` 角色負責本專案的架構、資料關係、操作流程與狀態圖。協調者在圖表有助於釐清設計或說明變更時主動選用，明確指派文件／圖表 hunks；小任務可直接採用該角色指引，不必為每張圖啟動代理。角色註冊與 MCP 連線以 `config.toml` 為準，職責細節以 `agents/mermaid.toml` 為準；不在應用程式加入 Mermaid runtime。

### 專案適用範圍

| 圖表目的 | 必讀來源與邊界 |
| --- | --- |
| 儲存庫與專案關係 | [Project](../docs/features/project/projects.md)：Space 下平行業務，專案引用單一來源儲存庫，引用不增加來源權限 |
| 物件與子工序 | [Object](../docs/features/repository/objects.md)、[Subobject](../docs/features/repository/subobjects.md)：單層子項，數量／單位／進度獨立 |
| 問題與討論 | [Issue](../docs/features/repository/issues.md)、[Discussion](../docs/features/repository/discussions.md)：各自狀態與資料，不自動推進物件進度 |
| 授權與寫入流程 | [操作模型](../docs/foundation/architecture/business-operation.md)：區分主體、資源、權限、領域規則、交易與結果 |

目前 `SpaceID` 對應 LINE 群組；實體位置 Space 的新設計必須標示「提案」，不能混入現行資料模型。圖表須註明來源版本、相關未提交變更及已實作／提案／待確認狀態，不以舊圖譜或規劃文件代替現行程式證據。

### 操作與驗收

1. 核對來源並在指派的 Markdown 保存可編修 `mermaid` 區塊及相對來源連結。
2. 啟動 Codex 的程序須繼承 `MERMAID_TOKEN`；token 值只由主機環境提供，不放入儲存庫、mise 或圖表。設定使用環境變數映射的 Authorization 標頭。
3. 確認主機實際曝光 Mermaid 工具並讀取 schema，再以 `validate_and_render_mermaid_diagram` 驗證最小脫敏圖表；送往遠端的圖表不得包含憑證、真實群組 ID、私人紀錄或整包程式碼。主機提供 Mermaid Chart 時，可用 display_mermaid 預覽並註明實際工具。專案 MCP 設定預設只開放渲染、標題與摘要工具，不開放帳號圖表寫入。
4. 核對圖中文字與箭頭符合來源，另執行設定檢查與 `git diff --check`。記錄實際渲染結果；未曝光工具、缺憑證或連線失敗時保留本機草稿並回報，不宣稱已驗證。

已執行中的聊天不會因修改設定自動取得工具；須由載入本專案設定且繼承環境變數的新 Codex 工作階段驗收。設定解析、主機載入、遠端渲染與業務語意為不同完成條件。

設定依據：[Mermaid MCP 官方文件](https://mermaid.ai/docs/ai/mcp-server) 與 [Codex 設定參考](https://developers.openai.com/codex/config-reference)，2026-11-18 查核；官方頁面未固定版本。圖表語法依當次任務查證，不假定遠端 renderer 與本機版本相同。
