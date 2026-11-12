# MINI App 前端

## 目錄導覽與維護

miniapp/ 為 Next.js／React workspace；src/pages/ 提供路由，business/attendance/ 提供出勤畫面，business/form/journal/ 提供日誌表單，identity/、platform/line/、shared/、shell/ 分別處理身分、平台、共用與外框。

進入 miniapp/ 先讀其 AGENTS.md 與安裝版本的 Next.js 指引。HTTP 呼叫透過 @repo/api-client；位置由瀏覽器取得，後端驗證。不手改建置輸出。

共用 UI primitives 位於 `miniapp/src/shared/ui/`，樣式工具位於 `shared/utils.ts`；業務畫面留在 `business/`。`components.json` 與 `tsconfig.json` 管理生成路徑及別名。Tailwind 由 Pages Router `_app.tsx` 經 PostCSS 編譯 `styles/globals.css`，包含 Preflight。頁面排版使用 utilities，元件外觀由 shadcn variants 與主題變數控制；不再複製或載入獨立的舊 CSS。

`shell/AppFrame.tsx` 套用官方 [sidebar-07 整合](miniapp/src/shell/sidebar-07/README.md)，統一側欄、頁首與手機開關；`Shell.tsx` 負責正式業務頁的身分及入口生命週期。空間預覽與定位頁重用 AppFrame；頁面不再各自維護導覽。

從根目錄透過 tool-exec.sh 執行 pnpm --filter @repo/miniapp check 與相關既有 test；需要靜態輸出時執行 build。API 變更另走 ../api/README.md。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [文件索引：Frontend](../docs/README.md#按模組閱讀)
- [Go browser Session 與滑動續期](../docs/architecture/browser-session.md)
- [React 架構與生命週期](../docs/architecture/react-frontend.md)
- [shadcn UI 與 MCP：元件位置、CSS 與工具驗證](../docs/tooling/mcp/shadcn.md)
- [瀏覽器定位整合](../docs/integrations/web-geolocation.md)
- [出勤查詢與更正](../docs/features/attendance-queries.md)
- [API 契約工作流程](../api/README.md)

- [群組請假登記](../docs/features/leave.md)

- [Business 模組重構](../docs/development/business-refactor.md)

## Business 畫面範圍

| 業務 | 既有前端實作 | 其他入口 |
| --- | --- | --- |
| Form / Journal | src/business/form/journal | [日誌 Google 表單](../docs/features/journal.md) |
| Attendance | src/business/attendance/{punch,records,corrections,approvals,leave} | Worktime 由紀錄畫面呈現；Board 透過 LINE |
| Summary | 沒有既有 MINI App 畫面 | [訊息收集、摘要指令與內部工作](../api/contract/summary/README.md) |
| Weather / Forecast | 沒有既有 MINI App 畫面 | [LINE 群組位置查詢](../api/contract/weather/README.md) |

不建立空畫面或讓瀏覽器呼叫內部摘要 worker。新增摘要／天氣網頁及公開 API 必須另定操作契約。

- [群組 Repository](../docs/features/repository.md)

## shadcn 頁內組合

補卡確認使用 Checkbox；空間預覽的清單與看板動作使用 Button、狀態篩選使用單選 ToggleGroup。PagePanel 沿用 Card 預設圓角與內距。Checkbox、Toggle、ToggleGroup 透過固定版 CLI 的 registry 輸出安裝，只修正專案 alias，沿用既有 radix-ui 依賴。sidebar-07 的 asChild 原生 button 保留，屬於官方元件組合。審核清單同步行為見 [出勤查詢與更正](../docs/features/attendance-queries.md#審核畫面狀態同步)。

出勤明細使用 Table；補卡／審核申請、請假紀錄與空間清單使用 Item；空間預覽搜尋使用 InputGroup。空白清單在成功查詢後以 Empty 呈現，載入與查詢失敗維持獨立狀態。安裝來源及用途見 [shadcn 指引](../docs/tooling/mcp/shadcn.md#頁面元件應用)。
