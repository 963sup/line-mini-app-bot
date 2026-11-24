# Codex CLI 操作

專案採單一代理，能力選擇依[環境入口](../../../../README.md#依任務選擇能力)。環境建立與更新由[Cloud 手冊](../../environment/codex-cloud.md)維護。

## 開啟專案與提交 Cloud 任務

在已建立的 Cloud 工作區，先確認 checkout 位於 `/workspace/line-mini-app-bot`。若需在該 checkout 開啟 Codex CLI，使用 `codex -C /workspace/line-mini-app-bot`。從工作區外提交 Cloud 任務時，先用 `codex cloud` 查環境 ID，再用 `codex cloud exec --env <ENV_ID> "<任務>"`。路徑用於選擇本機工作目錄，Cloud 任務由環境 ID 選擇。命令語法可用本機 CLI 的 `--help` 核對。

## 從根目錄處理跨模組任務

從[根目錄導覽](../../../../README.md#巢狀目錄導覽)找到目標模組及其工作指引。

## 專案設定與驗證

Skills 與 MCP 由 [.codex/config.toml](../../../../.codex/config.toml) 維護；工具版本與任務由 [mise.toml](../../../../mise.toml) 維護。設定影響後續載入此 checkout 的工作階段；目前聊天的工具曝光仍須依[載入檢查](../../environment/troubleshooting.md#設定與-mcp-診斷)核對。執行 `bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py` 驗證專案設定及能力指引。

VS Code 從儲存庫根目錄開啟，使用「Tasks: Run Task」執行[工作區任務](../../../../.vscode/tasks.json)：Go 環境準備、環境診斷、Go 應用檢查與儲存庫設定檢查。

## Codex 與 OpenAI 文件查證

查 Codex agent、MCP、設定或 OpenAI API 行為時，使用目前聊天可用的 `openaiDeveloperDocs` `search_openai_docs`、`fetch_openai_doc`。此遠端服務需要網路；網站回應 403 不代表 MCP 無法使用。工具未曝光時，依[MCP 載入診斷](../../environment/troubleshooting.md#設定與-mcp-診斷)處理，不把服務設定或安裝視為聊天已載入。

主機提供的 MCP 與權限仍適用；可選服務失敗時記錄實際限制。設定語法依[官方 MCP 文件](https://developers.openai.com/codex/mcp)查證。
