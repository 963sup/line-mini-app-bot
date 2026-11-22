# Codex 專案配置

專案使用單一代理，依任務讀取能力指引，不註冊或啟動子代理。

## 設定入口

| 內容 | 權威來源 |
| --- | --- |
| 模型功能、Skills 與 MCP | [config.toml](config.toml) |
| 探索、設計、實作、審查、測試與文件能力 | [能力入口](../docs/development/workflow/capabilities/README.md) |
| 任務規劃與驗收 | [工作流程](../docs/development/workflow/task-start.md) |
| 工具版本與任務 | [mise.toml](../mise.toml) |

## 初始化與載入

從實際 checkout 開啟 Codex；環境準備見[環境手冊](../docs/development/environment/README.md)。專案設定影響後續載入此 checkout 的工作階段，不會改寫既有聊天的主機權限或已啟動代理。

### Codex Cloud 新環境驗收

依[新環境驗收](../docs/development/environment/codex-cloud.md#新環境驗收)確認設定；MCP 曝光依[主機載入檢查](../docs/development/tooling/agents/README.md#project-loading-check)。

## 驗證

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py
```

檢查要求停用 multi_agent、沒有子代理註冊且能力指引齊全。Repomix 操作見[共用指南](../docs/development/tooling/shared/repomix.md)，文件圖見[Graphify](../docs/development/tooling/shared/graphify.md#docs-only-graph)。

[本目錄指引](AGENTS.md) · [根目錄指引](../AGENTS.md)
