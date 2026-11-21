# VS Code 設定維護

本目錄沿用[根目錄指引](../AGENTS.md)。任務以儲存庫根目錄為工作目錄，Go 與 gopls 使用專案的工具啟動腳本。工具版本由 [mise.toml](../mise.toml) 管理，維持編輯器工具自動更新關閉。共用設定不得包含個人路徑或憑證。

修改後從儲存庫根目錄執行 `python3 scripts/check/check-config.py`，並核對任務引用的腳本路徑。任務名稱或用途變更時，同步更新[本層說明](README.md)。
