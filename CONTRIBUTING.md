# 貢獻指南

1. 直接在 `main` 開發，開始前核對遠端版本及工作樹。
2. 更新相關文件，依[工程決策與交付要求](AGENTS.md#工程決策與交付要求)定義驗收並處理回歸證據缺口。
3. 依變更範圍執行[驗證流程](docs/workflows/verification.md)，實際結果、失敗與證據缺口須如實回報。
4. 每次提交與推送都須通過[發布流程](docs/workflows/release.md)的階段門檻。

請勿提交機密、工具快取或產生的分析輸出。安全漏洞請依 [安全政策](SECURITY.md) 私下回報。

## 開發與驗證

```bash
pnpm install --frozen-lockfile
go -C backend test ./...
go -C backend vet ./...
pnpm --filter @repo/miniapp check
pnpm --filter @repo/miniapp test
git diff --check
```

實際命令以目前 manifest、Go module 與 CI workflow 為準。格式修改只限本次擁有的檔案，勿順手重排其他內容。

## 分析與提交

按需使用 scoped search、程式圖或 Repomix bundle 定位關係，再回到目前原始碼與契約核對。工具無法解析的格式與排除範圍須明確列出；不得為補覆蓋率而繞過敏感資料排除。`.dockerignore`／`.gcloudignore` 只保留實際建置輸入，不因分析需求擴大上傳範圍。

提交前檢查 staged diff 與產物排除，保留其他人的本機修改。
交付時回報實際命令、失敗及未驗證整合；安裝或 MCP 握手不等於雲端部署或裝置驗收。
產生的分析包、圖譜及 profile 留在工作區，分享前依 [安全政策](SECURITY.md) 檢查敏感內容。
