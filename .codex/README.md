# Codex 專案配置

## 目錄導覽與維護

config.toml 定義角色與 MCP 連線，agents/ 定義角色指引，hooks.json 保存 hooks；skills/ 為安裝產物。

連線設定只在 config.toml 維護；工具版本、任務與非機密環境以 ../mise.toml 為準。不要手改上游技能或寫入憑證。

從根目錄執行 bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py；MCP transport 與聊天主機曝光需分別驗證。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)
