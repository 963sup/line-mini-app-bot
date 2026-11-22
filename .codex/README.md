# Codex 專案配置

## Codex 環境初始化

從儲存庫根目錄開啟 Codex。依[環境手冊](../docs/development/environment/README.md)選擇就地開發或 Cloud，Install script、Start skill 與快照更新見 [Codex Cloud 環境](../docs/development/environment/codex-cloud.md)。

## 用途與範圍

保存 Codex 角色與 MCP 設定。專案不使用 Codex hooks。

## 架構與責任邊界

此處定義專案的代理配置，工具版本與任務由根目錄 mise 設定管理。

## 閱讀與修改入口

先查閱[config.toml](config.toml)、[角色設定](agents/README.md)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[操作模型](../docs/contracts/business-operations.md) · [符合性流程](../docs/development/workflow/implementation.md#business-operation-compliance)

## 目錄導覽與維護

config.toml 定義角色與 MCP 連線，agents/ 定義角色指引；skills/ 為安裝產物。

連線設定只在 config.toml 維護；工具版本、任務與非機密環境以 ../mise.toml 為準。不要手改上游技能或寫入憑證。

從根目錄執行 bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 確認專案設定載入

從儲存庫根目錄開啟 Codex，以新工作階段實際列出的角色確認專案設定。MCP 依[主機載入檢查](../docs/development/tooling/agents/README.md#project-loading-check)驗收。[官方設定層級](https://developers.openai.com/codex/config-basic/)

Repomix 技能與 Graphify 文件圖依任務按需產生，操作見 [Repomix Skills](../docs/development/tooling/agents/repomix-skills.md#operation)與 [Graphify 文件圖](../docs/development/tooling/agents/graphify.md#docs-only-graph)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow/implementation.md#business-operation-compliance)。

現有角色依 Business Operation 契約分工；角色開始相關任務時須讀取模型及符合性流程，文件路徑不是自動 include。

## 角色配置與派工

角色檔位於 `agents/`，由 `config.toml` 註冊；完整職責、派工順序、深度與單一寫入者規則見[代理工作流程](../docs/development/workflow/task-start.md#delegation-depth-and-role-selection)。探索與查證由 explorer 承接，開發與除錯由 implementer 承接，文件、圖表與授權的 Wiki 維護由 documentation 承接。工具操作依[MCP 指南](../docs/development/tooling/agents/README.md)，不在角色檔重複連線設定。

支援自訂代理的 Codex 可選用已註冊角色；其他主機依工作流程採用明示的 fallback。配置存在不代表目前聊天主機已載入角色或 MCP。角色沿用父代理權限，不增加推送、發布或部署授權。

## Mermaid

`documentation` 角色負責本專案的架構、資料關係、操作流程與狀態圖。協調者在圖表有助於釐清設計或說明變更時主動選用，明確指派文件／圖表 hunks；小任務可直接採用該角色指引，不必為每張圖啟動代理。角色註冊與 MCP 連線以 `config.toml` 為準，職責細節以 `agents/documentation.toml` 為準；不在應用程式加入 Mermaid runtime。

安裝、領域邊界與渲染驗收見 [Mermaid 工具指南](../docs/development/tooling/shared/mermaid.md)。

### Codex Cloud 新環境驗收

從新聊天執行[新環境驗收](../docs/development/environment/codex-cloud.md#新環境驗收)。
