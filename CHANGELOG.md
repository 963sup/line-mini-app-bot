# Changelog

此檔記錄有意義的儲存庫變更；Unreleased 不代表已發布或部署。

## [Unreleased]

### Added

- Go LINE 平台的個人／群組架構、實作契約、命名與驗收文件。
- Web Geolocation API 權限、資料驗證與實機測試規範。
- 專案 skills、來源紀錄、Samber skills lockfile 與 skills CLI 工具管理。
- Graphify／Repomix 專用排除規則、Git attributes 與安全通報政策。

### Changed

- 應用部署目標統一為 Cloud Run；Worker 與 Google ADC/IAM 仍待實作驗收。
- Repomix 採用 `.repomixignore`，XML 輸出選項由 mise 任務管理，不再維護 JSON 配置。
- Graphify／Repomix 與部署 ignore 明確排除 `.agents`、`.claude`、`agent` 及 `skills-lock.json`；分析保留 `.codex`。
- 根目錄文件補齊 Codex Cloud 工具入口、分析限制及產物處理；Go EditorConfig 明確設定 tab 寬度。

### Implementation status

- 現有 HTTP 服務已組裝 LINE Webhook、MINI App、出勤及摘要模組，沒有健康檢查路由；真實服務、裝置及部署驗收仍須另行確認。
- 使用者層 Codex plugins 不包含在 Git；安裝紀錄見 docs/tooling/skills.md。
