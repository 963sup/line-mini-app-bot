# Changelog

## 2026-11-11 補卡與審核閉環

- 沿用更正核心，支援無班次補卡與既有班次時間更正；一律人工核准後才生效。
- 新增本人／授權群組申請清單、駁回理由及本人撤回，保留白名單、禁止自審與原始打卡證據。
- 群組快捷入口改為「補卡」；既有「漏下班」仍可查未結束班次，「現在下班」維持真實目前時間。
- 首版補卡申報完整已結束班次，同員工／群組重疊檢查上限1000筆；完整權限管理與仍在工作的上班補登另行擴充。

此檔記錄有意義的儲存庫變更；Unreleased 不代表已發布或部署。

## [Unreleased]

### Added

- MINI App Cookie Session、滑動續期與伺服器登出撤銷。
- 群組請假登記、出勤人數看板與本人漏下班檢查。
- 出勤月紀錄、更正與審核的 React 操作流程。

- Go LINE 平台的個人／群組架構、實作契約、命名與驗收文件。
- Web Geolocation API 權限、資料驗證與實機測試規範。
- 專案 skills、來源紀錄、Samber skills lockfile 與 skills CLI 工具管理。
- Graphify／Repomix 專用排除規則、Git attributes 與安全通報政策。

### Changed

- LINE 群組選單加入出勤、請假與漏下班入口，簡化出勤選單與移除連線狀態。
- Rich Menu 容器素材統一至 `/assets/line/richmenu`，並整理 nonroot 讀取權限。
- 文件同步 Session、React 與部署設定，補齊故障排查、歷史快照標示及舊版本清理規則。

- 應用部署目標統一為 Cloud Run；Cloud Tasks 摘要入口已有實作，常駐 Worker 不在目前範圍。ADC／IAM 與真實整合驗收依目標環境另行確認。
- Repomix 採用 `.repomixignore`，XML 輸出選項由 mise 任務管理，不再維護 JSON 配置。
- Graphify／Repomix 與部署 ignore 明確排除 `.agents`、`.claude`、`agent` 及 `skills-lock.json`；分析保留 `.codex`。
- 根目錄文件補齊 Codex Cloud 工具入口、分析限制及產物處理；Go EditorConfig 明確設定 tab 寬度。

### Implementation status

- 現有 HTTP 服務已組裝 LINE Webhook、MINI App、出勤及摘要模組，沒有健康檢查路由；真實服務、裝置及部署驗收仍須另行確認。
- 使用者層 Codex plugins 不包含在 Git；安裝紀錄見 docs/tooling/skills.md。
