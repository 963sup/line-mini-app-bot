# GitHub 設定維護

本目錄沿用[根目錄指引](../AGENTS.md)。協作範本與 CI 入口見[本層說明](README.md)。

修改 workflow 時保留最小權限，第三方 action 固定到提交 SHA。若更動腳本路徑，核對[腳本索引](../scripts/README.md)，並從儲存庫根目錄執行 `python3 scripts/check/check-config.py` 與 `git diff --check`。
