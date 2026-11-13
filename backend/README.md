# Go 後端

Space及Project各自保存核心、用例與adapters。Project不依賴Repository核心模型作父容器，只透過使用端定義的來源讀取port查事項；Space成員及來源資格分開核驗。現行邊界與遷移見 [獨立Project](../docs/features/projects.md)。

來源Issue hydration故障與Project自有資料錯誤的隔離、回歸及架構度量見[第二輪收斂證據](../docs/tooling/architecture-convergence-pass2-20261113.md)。

`cmd/projectmigrate` 是明確namespace的dry-run／apply／受限rollback工具，旗標與writer部署前提見 [遷移指南](../docs/operations/project-migration.md)；不由server啟動時自動搬正式資料。

## 目錄導覽與維護

cmd/ 保存 server、richmenu、firestorecheck 與 attendancemigrate 入口；internal/business/ 按 attendance、summary、weather 及其子功能分工；identity 等共用能力保留於 internal/。httpapi/ 保存契約接線與產生碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/server。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依 ../docs/agents/acceptance.md 執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [日誌表單入口](../docs/features/journal.md)

- [文件索引：Backend](../docs/README.md#按模組閱讀)
- [Go browser Session 與滑動續期](../docs/architecture/browser-session.md)
- [後端功能邊界](../docs/architecture/backend-boundaries.md)
- [依賴規則](../docs/architecture/dependencies.md)
- [Go 開發慣例](../docs/development/go.md)
- [可靠性與資料](../docs/architecture/reliability.md)
- [API 契約工作流程](../api/README.md)

- [群組請假登記](../docs/features/leave.md)

- [群組出勤看板與本人漏下班檢查](../docs/features/attendance-board.md)

- [Business 模組重構](../docs/development/business-refactor.md)

Summary 的共用 LINE 組合入口在 business/summary/line；訊息收集位於
messagecollection/line，摘要指令與排程位於 summarization/line。
收集、指令、worker 使用各自需要的 Store 介面，共用 Firestore 交易保留在 summary/firestore。
Weather 既有實作集中在 business/weather/forecast/{cwa,line}。
入口契約見 [Summary](../api/contract/summary/README.md) 與 [Weather](../api/contract/weather/README.md)。

- [群組 Repository](../docs/features/repository.md)
- [基礎雛形與停止擴充界線](../docs/features/repository-prototype.md)
- [空間標籤目錄](../docs/features/labels.md)：沿用 Repository client，建立名稱綁定及防重收據，不占共用編號。
- [空間討論與留言](../docs/features/discussions.md)：獨立核心及 HTTP／Firestore adapters，重用既有 client 與編號。
- [空間事項第一階段](../docs/features/issues.md)：核心、HTTP 與 Firestore 分層實作；正式環境待驗收。
- [Repository 共用資源編號基礎](../docs/features/repository-numbering.md)：內部交易介面，已由事項建立使用，不提供獨立公開分配 API。

- [里程碑、專案與統計](../docs/features/planning.md)

## 協作能力分佈與交付狀態

本機新增的 Issue、Discussion、Project 及 Repository labels／milestones／statistics 分工、HTTP／server 接線與資料 adapter 位置見 [能力分佈](../docs/development/business-refactor.md#協作能力的本機分佈2026-11-13)。該表區分原碼接線、尚未提交檔案、既有驗證紀錄與待確認的正式驗收；本次文件更新不代表新增模組已發布或重新通過 Go 檢查。
