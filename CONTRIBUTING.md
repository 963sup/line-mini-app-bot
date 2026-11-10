# 貢獻指南

先閱讀 [文件索引](docs/README.md)、[環境設定](docs/development/setup.md) 與 [Go 慣例](docs/development/go.md)。

1. 從最新目標分支建立工作分支，讓每次變更聚焦於單一問題。
2. 更新相關文件；目前依 [已接受的基線](docs/architecture/decisions/0002-baseline.md) 先穩定基礎，不新增測試方案、案例或工具。
3. 依下列步驟執行設定、Shell 與格式檢查；Go 修改另跑既有 CI 對應任務。
4. 現有 Go 應用與 CI 保留；[測試指引](docs/development/testing.md) 作為後續階段參考，實際執行結果應如實回報。
5. 提出 Pull Request，說明問題、修改與實際驗證結果。

請勿提交機密、工具快取或產生的分析輸出。安全漏洞請依 [安全政策](.github/SECURITY.md) 私下回報。

## Cloud 開發與驗證

先執行 `bash scripts/tooling/bootstrap.sh core`；分析／MCP 工具使用 `all`。
日常命令透過 `scripts/tooling/tool-exec.sh` 載入 `mise.toml` 的版本與環境，不另改 HOME。

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py
bash scripts/tooling/tool-exec.sh mise run analysis-scope
find scripts -type f -name '*.sh' -exec bash -n {} +
git diff --check
```

Go 修改依 [驗收規則](docs/agents/acceptance.md) 執行 focused checks，並完成既有 CI 的
`fmt-check`、`lint`、`check`、`build` 與 `vuln` 任務；不新增測試框架，也不跳過既有檢查。
例如：`bash scripts/tooling/tool-exec.sh mise run check`。格式修改先取得檔案所有權，勿順手重排其他檔案。

## 分析與提交

修改來源後跑 `mise run graphify-update`、`mise run repomix`，同樣透過 `scripts/tooling/tool-exec.sh`。
用圖譜定位程式關係，再查原始碼與 Repomix 原文；AST 不支援的格式須明確列出。
`.graphifyignore`／`.repomixignore` 排除 `.agents`、`.claude`、`agent` 與 `skills-lock.json`；
不得為補覆蓋率而繞過這些排除。`.dockerignore`／`.gcloudignore` 只保留實際建置輸入，
不因分析需求擴大上傳範圍。根目錄 `analysis-scope.toml` 是專案自訂分類契約，
不是 Graphify／Repomix 原生設定；獨立 audit 與日常 CLI 範圍須分別核對。

提交前檢查 staged diff 與產物排除，保留其他人的本機修改。
PR 回報實際命令、失敗及未驗證整合；安裝或 MCP 握手不等於雲端部署或裝置驗收。
產生的分析包、圖譜及 profile 留在工作區，分享前依 [安全政策](SECURITY.md) 檢查敏感內容。
