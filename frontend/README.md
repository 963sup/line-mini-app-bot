# MINI App 前端

## 目錄導覽與維護

miniapp/ 為 Next.js／React workspace；src/pages/ 提供路由，features/attendance/ 提供出勤畫面，identity/、platform/line/、shared/、shell/ 分別處理身分、平台、共用與外框。

進入 miniapp/ 先讀其 AGENTS.md 與安裝版本的 Next.js 指引。HTTP 呼叫透過 @repo/api-client；位置由瀏覽器取得，後端驗證。不手改建置輸出。

從根目錄透過 tool-exec.sh 執行 pnpm --filter @repo/miniapp check 與相關既有 test；需要靜態輸出時執行 build。API 變更另走 ../api/README.md。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [文件索引：Frontend](../docs/README.md#按模組閱讀)
- [Go browser Session 與滑動續期](../docs/architecture/browser-session.md)
- [React 架構與生命週期](../docs/architecture/react-frontend.md)
- [瀏覽器定位整合](../docs/integrations/web-geolocation.md)
- [出勤查詢與更正](../docs/features/attendance-queries.md)
- [API 契約工作流程](../api/README.md)
