# 貢獻指南

先閱讀 [文件索引](docs/README.md)、[環境設定](docs/development/setup.md) 與 [Go 慣例](docs/development/go.md)。

1. 從最新目標分支建立工作分支，讓每次變更聚焦於單一問題。
2. 更新相關文件；目前依 [已接受的基線](docs/architecture/decisions/0002-baseline.md) 先穩定基礎，不新增測試方案、案例或工具。
3. 執行 `python3 scripts/check-config.py`、`bash -n scripts/*.sh` 與 `git diff --check`。
4. 現有 Go 應用與 CI 保留；[測試指引](docs/development/testing.md) 作為後續階段參考，實際執行結果應如實回報。
5. 提出 Pull Request，說明問題、修改與實際驗證結果。

請勿提交機密、工具快取或產生的分析輸出。安全漏洞請依 [安全政策](.github/SECURITY.md) 私下回報。
