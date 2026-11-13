# Codex 專案配置

## 目錄導覽與維護

config.toml 定義角色與 MCP 連線，agents/ 定義角色指引，hooks.json 保存 hooks；skills/ 為安裝產物。

連線設定只在 config.toml 維護；工具版本、任務與非機密環境以 ../mise.toml 為準。不要手改上游技能或寫入憑證。

從根目錄執行 bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py；MCP transport 與聊天主機曝光需分別驗證。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/business-operation.md) 與 [符合性流程](../docs/agents/workflow.md#business-operation-compliance)。

現有角色依 Business Operation 契約分工；角色開始相關任務時須讀取模型及符合性流程，文件路徑不是自動 include。

## 文件代理

`documentation` 負責指派範圍內的 Markdown 編寫、審查與修正，依[文件規範](../docs/development/documentation.md)執行；協調者分派單一寫入者並整合結果。角色檔位於 `agents/documentation.toml`，由 `config.toml` 註冊，沿用父代理模型、權限與 MCP，不增加推送或部署授權。

支援自訂代理的 Codex 可選用此角色；其他主機由協調者傳遞角色指引或自行承擔職責。配置存在不代表目前聊天主機已載入或可原生選用該角色，須分別核對主機工具能力及實際執行結果。
