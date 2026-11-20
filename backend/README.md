# Go 後端

## 用途與範圍

保存 Go 後端指令、套件與模組設定。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[package.json](package.json)、[turbo.json](turbo.json)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[獨立Project](../docs/domains/project/project.md) · [工程日誌與群組發布](../docs/domains/journal/journal.md)


工程群組及Project各自保存核心、用例與adapters。Project不依賴Repository核心模型作父容器，只透過使用端定義的來源讀取port查事項；工程群組成員及來源資格分開核驗。現行邊界與遷移見 [獨立Project](../docs/domains/project/project.md)。

`cmd/projectmigrate` 是明確namespace的dry-run／apply／受限rollback工具，旗標與writer部署前提見 [遷移指南](cmd/projectmigrate/README.md)；不由server啟動時自動搬正式資料。

## 目錄導覽與維護

可執行入口位於 [cmd](cmd)。[business](internal/business) 依業務領域分工，[bootstrap](internal/bootstrap) 組裝服務，[identity](internal/identity) 處理身分與 Session，[httpapi](internal/transport/http/api) 保存 OpenAPI 契約接線與生成碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/bootstrap。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依 ../docs/development/acceptance.md 執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [工程日誌與群組發布](../docs/domains/journal/journal.md)

- [文件索引：Backend](../docs/README.md#按模組閱讀)
- [Go browser Session 與滑動續期](../docs/architecture/identity/browser-session.md)
- [後端功能邊界](../docs/architecture/runtime/backend-boundaries.md)
- [依賴規則](../docs/architecture/runtime/dependencies.md)
- [Go 開發慣例](../docs/standards/go.md)
- [可靠性與資料](../docs/proposals/webhook-delivery/design.md)
- [API 契約工作流程](../api/README.md)

- [群組請假登記](../docs/domains/attendance/leave.md)

- [群組出勤看板與本人漏下班檢查](../docs/domains/attendance/board.md)

- [目前模組與責任](../docs/architecture/runtime/code-layout.md)

摘要的 [LINE 入口](internal/business/summary/adapter/line) 串接訊息收集與摘要指令。[Cloud Tasks 適配器](internal/business/summary/adapter/messaging/cloudtasks/summarization) 排程工作，[Firestore 適配器](internal/business/summary/adapter/firestore) 保存訊息與工作狀態。
天氣預報由 [CWA 適配器](internal/business/weather/adapter/cwa/forecast) 取得資料，再由 [LINE 適配器](internal/business/weather/adapter/line/forecast) 組裝回覆。
入口契約見 [Summary](../api/contract/summary/README.md) 與 [Weather](../api/contract/weather/README.md)。

- [群組 Repository](../docs/domains/repository/README.md)
- [基礎雛形與停止擴充界線](../docs/domains/repository/acceptance.md)
- [空間標籤目錄](../docs/domains/repository/labels.md)：沿用 Repository client，建立名稱綁定及防重收據，不占共用編號。
- [空間討論與留言](../docs/domains/repository/discussions.md)：獨立核心及 HTTP／Firestore adapters，重用既有 client 與編號。
- [空間事項第一階段](../docs/domains/repository/issues.md)：核心、HTTP 與 Firestore 分層實作；正式環境待驗收。
- [Repository 共用資源編號基礎](../docs/domains/repository/numbering.md)：內部交易介面，已由事項建立使用，不提供獨立公開分配 API。

- [里程碑](../docs/domains/repository/milestones.md)、[Project](../docs/domains/project/project.md)與[洞察（Insights）](../docs/domains/repository/insights.md)

## 協作能力分佈與交付狀態

業務模組位於 [internal/business](internal/business)，服務組裝見 [bootstrap](internal/bootstrap)，各功能行為見[功能索引](../docs/domains/README.md)。

[現行實作圖：格式與資料相容](../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../docs/diagrams.md#未來設計)
