# MINI App 前端

專案與儲存庫在側欄同層。專案頁自行管理工程群組與 Project 選取；來源選擇的讀取資格及舊入口轉接規則見 [Project 契約](../docs/features/project/projects.md)。

## 目錄導覽與維護

miniapp/ 為 Next.js／React workspace；src/pages/ 提供路由，business/attendance/ 提供出勤畫面，business/form/journal/ 提供日誌表單，identity/、platform/line/、shared/、shell/ 分別處理身分、平台、共用與外框。

進入 miniapp/ 先讀其 AGENTS.md 與安裝版本的 Next.js 指引。HTTP 呼叫透過 @repo/api-client；位置由瀏覽器取得，後端驗證。不手改建置輸出。

共用 UI primitives 位於 `miniapp/src/shared/ui/`，樣式工具位於 `shared/utils.ts`；業務畫面留在 `business/`。`components.json` 與 `tsconfig.json` 管理生成路徑及別名。Tailwind 由 Pages Router `_app.tsx` 經 PostCSS 編譯 `styles/globals.css`，包含 Preflight。頁面排版使用 utilities，元件外觀由 shadcn variants 與主題變數控制；不再複製或載入獨立的舊 CSS。

所有前端 UI 任務主動遵循 [官方元件保護與選用流程](miniapp/src/shared/ui/AGENTS.md)，先搜尋 [Components／AI Elements 本機參考庫](../docs/tooling/shadcn-reference/README.md)，再透過 shadcn Skill、可用的 MCP 或固定 CLI 核對並按需組合／安裝。既有官方元件禁止直接修改，業務包裝放在應用層。

`shell/AppFrame.tsx` 套用官方 [sidebar-07 整合](miniapp/src/shell/sidebar-07/README.md)，統一側欄、頁首與手機開關；`Shell.tsx` 負責正式業務頁的身分及入口生命週期。定位頁重用 AppFrame；頁面不再各自維護導覽。

從根目錄透過 tool-exec.sh 執行 pnpm --filter @repo/miniapp check 與相關既有 test；需要靜態輸出時執行 build。API 變更另走 ../api/README.md。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [正式前端入口與生命週期](../docs/foundation/frontend/react-frontend.md)
- [文件索引：Frontend](../docs/README.md#按模組閱讀)
- [Go browser Session 與滑動續期](../docs/foundation/identity/browser-session.md)
- [shadcn UI 與 MCP：元件位置、CSS 與工具驗證](../docs/tooling/mcp/shadcn.md)
- [瀏覽器定位整合](../docs/foundation/integrations/web-geolocation.md)
- [出勤查詢與更正](../docs/features/attendance/queries.md)
- [API 契約工作流程](../api/README.md)

- [群組請假登記](../docs/features/attendance/leave.md)

- [目前模組與責任](../docs/foundation/architecture/layout.md)

## Business 畫面範圍

| 業務 | 既有前端實作 | 其他入口 |
| --- | --- | --- |
| 日誌表單 | `/miniapp/journal` 開啟 Google 表單 | [表單設定](../docs/features/journal/google-forms.md) |
| 工程日誌 | `/miniapp/dashboard` 與 `/miniapp/feed` 共用 JournalWorkspace | [日誌設計](../docs/features/journal/design.md) |
| 出勤 | 打卡、紀錄、更正、審核、請假；工時由紀錄頁呈現 | [出勤功能](../docs/features/attendance/attendance.md) |
| Repository | `/miniapp/repositories` 提供庫內工區、物件與協作功能 | [儲存庫功能索引](../docs/features/repository/README.md) |
| Project | `/miniapp/projects` 提供獨立的跨庫專案頁 | [Project 契約](../docs/features/project/projects.md) |
| 摘要與天氣 | 使用 LINE 入口，摘要背景工作由內部端點處理 | [摘要](../api/contract/summary/README.md) · [天氣](../api/contract/weather/README.md) |

## shadcn 頁內組合

補卡確認使用 Checkbox。PagePanel 沿用 Card 預設圓角與內距。既有 Checkbox、Toggle、ToggleGroup 透過固定版 CLI 的 registry 輸出安裝，只修正專案 alias，沿用既有 radix-ui 依賴。sidebar-07 的 asChild 原生 button 保留，屬於官方元件組合。審核清單同步行為見 [出勤查詢與更正](../docs/features/attendance/queries.md#審核畫面狀態同步)。

出勤明細使用 Table；補卡／審核申請、請假紀錄與空間清單使用 Item。空白清單在成功查詢後以 Empty 呈現，載入與查詢失敗維持獨立狀態。安裝來源及用途見 [shadcn 指引](../docs/tooling/mcp/shadcn.md#頁面元件應用)。

[現行實作圖：頁面與身分生命週期](../docs/foundation/frontend/react-frontend.md#頁面與身分生命週期) · [未來設計圖](../docs/diagrams.md#未來設計)
