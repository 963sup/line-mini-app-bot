# .codex 工作指引

本目錄沿用[根目錄指引](../AGENTS.md)。[config.toml](config.toml)登記角色與 MCP；[角色目錄](agents/README.md)維護分工，[hooks.json](hooks.json)定義 hooks。環境設定見[本層說明](README.md)。

## 職責與邊界

連線設定只在 config.toml 維護；版本、任務與非機密環境依 [mise.toml](../mise.toml)。上游技能由安裝流程管理，憑證由執行環境提供。

## 驗證

從根目錄執行 `bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py`。涉及 MCP 時，分別驗證 CLI 配置、transport 及目前聊天的原生工具。[驗證步驟](README.md#codex-cloud-新環境驗收)

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow.md#business-operation-compliance)。

角色調整必須同步引用操作模型與符合性流程，以角色責任補充，不複製完整政策；不藉由此流程增加角色或權限。
