# Go 後端

## 目錄導覽與維護

cmd/ 保存 server、richmenu、firestorecheck 與 attendancemigrate 入口；internal/business/ 按 attendance、summary、weather 及其子功能分工；identity 等共用能力保留於 internal/。httpapi/ 保存契約接線與產生碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/server。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依 ../docs/agents/acceptance.md 執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

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
