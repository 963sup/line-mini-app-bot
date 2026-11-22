# 專案技能維護指引

本目錄沿用[根目錄指引](../AGENTS.md)。修改技能來源前讀[目錄說明](README.md)與[技能維護流程](../docs/development/tooling/agents/skills.md)。

## 職責與邊界

上游技能保存於 `skills/`，來源與雜湊由[鎖檔](../skills-lock.json)維護。保留技能、來源與授權；專案規則集中在本庫指引。Go 的專案操作見[Go 開發規範](../docs/standards/go.md)與[驗收矩陣](../docs/development/tooling/testing/commands.md#變更驗收矩陣)，角色與 MCP 註冊見[Codex 設定](../.codex/config.toml)。

## 驗證

直接比對受影響的技能與鎖檔；引用打包或圖譜結果前，核對來源版本與實際涵蓋檔案。工具設定變更執行 `bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py`，角色及工具可用性另由當前主機驗證。
