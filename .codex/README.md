# Codex 專案配置

## 目錄導覽與維護

config.toml 定義角色與 MCP 連線，agents/ 定義角色指引，hooks.json 保存 hooks；skills/ 為安裝產物。

連線設定只在 config.toml 維護；工具版本、任務與非機密環境以 ../mise.toml 為準。不要手改上游技能或寫入憑證。

從根目錄執行 bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py；MCP transport 與聊天主機曝光需分別驗證。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/contracts/business-operations.md) 與 [符合性流程](../docs/agents/workflow.md#business-operation-compliance)。

現有角色依 Business Operation 契約分工；角色開始相關任務時須讀取模型及符合性流程，文件路徑不是自動 include。

## 角色配置與派工

角色檔位於 `agents/`，由 `config.toml` 註冊；完整職責、派工順序、深度與單一寫入者規則見[代理工作流程](../docs/agents/workflow.md#delegation-depth-and-role-selection)。`graphify`、`repomix` 提供探索證據，`context7` 查證適用版本文件，`golang`、`typescript`、`nextjs`、`shadcn` 接受明確範圍的實作指派，`documentation` 維護 Markdown。工具操作依[MCP 指南](../docs/tooling/mcp/README.md)，不在角色檔重複連線設定。

支援自訂代理的 Codex 可選用已註冊角色；其他主機依工作流程採用明示的 fallback。配置存在不代表目前聊天主機已載入角色或 MCP。角色沿用父代理權限，不增加推送、發布或部署授權。

## Mermaid

`mermaid` 角色負責本專案的架構、資料關係、操作流程與狀態圖。協調者在圖表有助於釐清設計或說明變更時主動選用，明確指派文件／圖表 hunks；小任務可直接採用該角色指引，不必為每張圖啟動代理。角色註冊與 MCP 連線以 `config.toml` 為準，職責細節以 `agents/mermaid.toml` 為準；不在應用程式加入 Mermaid runtime。

### 專案適用範圍

| 圖表目的 | 必讀來源與邊界 |
| --- | --- |
| 儲存庫與專案關係 | [Project](../docs/domains/project.md)：個人或群組擁有 Project，引用可跨庫，來源讀權逐庫核驗 |
| 物件與子工序 | [Object](../docs/domains/repository/objects.md)、[Subobject](../docs/domains/repository/objects.md)：單層子項，數量／單位／進度獨立 |
| 問題與討論 | [Issue](../docs/domains/repository/issues.md)、[Discussion](../docs/domains/repository/discussions.md)：各自狀態與資料，不自動推進物件進度 |
| 授權與寫入流程 | [操作模型](../docs/architecture/contracts/business-operations.md)：區分主體、資源、權限、領域規則、交易與結果 |

以 `GroupID` 識別 LINE 群組，以 `SpaceID` 識別 Repository 內的工區。[庫內工區](../docs/domains/repository/spaces.md)

圖表須註明來源版本、相關未提交變更及已實作／提案／待確認狀態，不以舊圖譜或規劃文件代替現行程式證據。

### 操作與驗收

1. 核對來源並依[文件圖表規範](../docs/standards/documentation.md#獨立圖檔)維護內嵌區塊或獨立圖檔。
2. 從專案根目錄安裝固定版 Mermaid v13 alpha，執行 `bash scripts/tooling/tool-exec.sh mise run install-mermaid`。
3. 執行 `bash scripts/tooling/tool-exec.sh mise run render-mermaid -- render diagram.mmd --format svg > diagram.svg`。
4. 核對圖中文字與箭頭符合來源，並記錄實際渲染結果。

專案預設使用本機渲染。v13 alpha 由 Node 直接渲染，無須 Chromium。遠端 Mermaid MCP 預設停用。用法見 [Mermaid v13 alpha](https://mermaid.js.org/intro/v13-alpha.html)。

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
