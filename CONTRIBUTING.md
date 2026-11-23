# 貢獻指南

先閱讀 [文件索引](docs/)、[環境設定](docs/development/tooling/shared/toolchain-setup.md) 與 [Go 慣例](docs/standards/go.md)。

1. 直接在 `main` 開發，開始前核對遠端版本及工作樹。
2. 更新相關文件，依[工程決策與交付要求](AGENTS.md#工程決策與交付要求)定義驗收並處理回歸證據缺口。
3. 依下列步驟執行設定、Shell 與格式檢查；Go 修改另跑既有 CI 對應任務。
4. 現有 Go 應用與 CI 保留；[驗收矩陣](docs/development/tooling/testing/commands.md#變更驗收矩陣)列出現階段應執行的檢查，實際執行結果應如實回報。
5. 每次提交與推送都須通過[發布流程](docs/development/workflow/publication.md#repository-publication)的階段門檻。

請勿提交機密、工具快取或產生的分析輸出。安全漏洞請依 [安全政策](.github/SECURITY.md) 私下回報。

## Cloud 開發與驗證

Codex Cloud 使用 `scripts/bootstrap.sh` 初始化。本機開發使用 `scripts/bootstrap.sh core`；MCP 工具按需另行安裝。
日常命令透過 `scripts/tooling/tool-exec.sh` 載入 `mise.toml` 的版本與環境，不另改 HOME。

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py
git diff --check
```

Go 修改依 [驗收規則](docs/development/tooling/testing/commands.md#變更驗收矩陣) 執行 focused checks，並完成既有 CI 的
`fmt-check`、`lint`、`check`、`build` 與 `vuln` 任務；不新增測試框架，也不跳過既有檢查。
例如：`bash scripts/tooling/tool-exec.sh mise run check`。格式修改先取得檔案所有權，勿順手重排其他檔案。

## 分析與提交

修改來源後跑 `mise run graphify-update`、`mise run repomix`，同樣透過 `scripts/tooling/tool-exec.sh`。
用圖譜定位程式關係，再查原始碼與 Repomix 原文；AST 不支援的格式須明確列出。
`.graphifyignore`／`.repomixignore` 排除 `.agents`、`.claude`、`agent` 與 `skills-lock.json`；
不得為補覆蓋率而繞過這些排除。`.dockerignore`／`.gcloudignore` 只保留實際建置輸入，
不因分析需求擴大上傳範圍。

提交前檢查 staged diff 與產物排除，保留其他人的本機修改。
交付時回報實際命令、失敗及未驗證整合；安裝或 MCP 握手不等於雲端部署或裝置驗收。
產生的分析包、圖譜及 profile 留在工作區，分享前依 [安全政策](SECURITY.md) 檢查敏感內容。
