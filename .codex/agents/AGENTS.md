# 角色設定維護指引

本目錄沿用[上層工作指引](../AGENTS.md)。調整角色名稱或檔案位置時，同步核對[註冊設定](../config.toml)中的角色名稱與 config_file。角色引用腳本時，從[腳本索引](../../scripts/README.md)確認入口。

從儲存庫根目錄執行 `python3 scripts/check/check-config.py`，驗證角色註冊、必要欄位與文件引用。目錄用途見[角色設定說明](README.md)。
