# GitHub 協作與 CI

回報錯誤或提出需求時使用 [Issue 範本](ISSUE_TEMPLATE/)。開 PR 時依 [PR 範本](pull_request_template.md) 說明變更與驗證結果。安全漏洞請依[安全政策](SECURITY.md)私下回報。

[GitHub Actions](workflows/check.yml) 在推送至 `main` 或更新 PR 時執行基本設定與文件檢查。純 Markdown 或 `docs/` 下的 Mermaid／SVG 變更會略過工具回歸、Go／前端／API 檢查與容器建置。只要包含其他檔案，或無法取得比較所需的歷史，就執行完整檢查。部署流程見[部署說明](../docs/operations/README.md)。

維護本目錄時請讀[工作指引](AGENTS.md)。
