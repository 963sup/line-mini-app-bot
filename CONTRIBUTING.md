# 貢獻指南

先閱讀 [文件索引](docs/README.md)、[環境設定](docs/development/setup.md) 與 [Go 慣例](docs/development/go.md)。

1. 從最新目標分支建立工作分支，讓每次變更聚焦於單一問題。
2. 更新相關文件，新增行為時提供合適的測試。
3. 執行 `python3 scripts/check-config.py`、`bash -n scripts/*.sh` 與 `git diff --check`。
4. Go 應用建立後，依 [測試指引](docs/development/testing.md) 執行相關檢查。
5. 提出 Pull Request，說明問題、修改與實際驗證結果。

請勿提交機密、工具快取或產生的分析輸出。安全漏洞請依 [安全政策](.github/SECURITY.md) 私下回報。
