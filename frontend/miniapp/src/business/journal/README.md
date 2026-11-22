# frontend/miniapp/src/business/journal 目錄導覽

組合工作日誌的工作區、輸入與查閱互動。

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

從 [GroupCharts.tsx](GroupCharts.tsx)、[WorkContext.tsx](WorkContext.tsx)、[Workspace.tsx](Workspace.tsx)、[Composer.tsx](Composer.tsx) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

- 側欄「快捷入口」：按「寫日誌」→ 選擇儲存庫 → 填寫並發布。工作牆依選取的儲存庫讀取日誌。
- Dashboard：依日期及儲存庫篩選，透過 [GroupCharts.tsx](GroupCharts.tsx) 為每個有日誌的儲存庫顯示一張 shadcn 長條圖，逐項呈現今日實際及明日預計累計完成度。載入更多時，同儲存庫的日誌會加到同一張圖表。展開日誌明細可查看人員及原文連結。
- Dashboard 圖表採用 shadcn 官方 [Bar](https://ui.shadcn.com/charts/bar#charts) 與 [Tooltip](https://ui.shadcn.com/charts/tooltip#charts) 的組合，提示顯示完整工作名稱和百分比。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[上層導覽](../README.md)

## 工作選項載入

同一份日誌草稿內，相同儲存庫的工作列共用物件、子物件及空間的載入結果與進行中請求。失敗移除該筆結果，重試重新請求；建立空間後使該儲存庫的選項失效。關閉草稿或切換身分時取消請求並清除資料，不跨草稿或使用者保存。現行 API 仍須逐物件查詢子物件，尚未改成後端搜尋介面。

工作牆與撰寫表單分開載入；只有開啟撰寫流程才下載 Composer 與其表單相依。圖表仍按原本方式延遲載入。
