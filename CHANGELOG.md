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
- Repomix 採用 .repomixignore，XML 輸出選項由 mise 任務管理，不再維護 JSON 配置。

### Implementation status

- 現有應用仍以 HTTP health server 為主；LINE 業務、MINI App 與持久 Worker 的設計不代表功能已完成。
- 使用者層 Codex plugins 不包含在 Git；安裝紀錄見 docs/tooling/skills.md。
