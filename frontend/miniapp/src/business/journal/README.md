# frontend/miniapp/src/business/journal 目錄導覽

## 用途與範圍

組合工作日誌的工作區、輸入與查閱互動。

## 架構與責任邊界

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[GroupCharts.tsx](GroupCharts.tsx)、[WorkContext.tsx](WorkContext.tsx)、[Workspace.tsx](Workspace.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

從 [Workspace.tsx](Workspace.tsx) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

- 側欄「快捷入口」：按「寫日誌」→ 選擇發布群組 → 填寫並發布。工作牆的群組篩選與發布群組分開。
- Dashboard：依日期及群組篩選，透過 [GroupCharts.tsx](GroupCharts.tsx) 為每個有日誌的群組顯示一張 shadcn 長條圖，逐項呈現今日實際及明日預計累計完成度。載入更多時，同群組的日誌會加到同一張圖表。展開日誌明細可查看人員及原文連結。
- Dashboard 圖表採用 shadcn 官方 [Bar](https://ui.shadcn.com/charts/bar#charts) 與 [Tooltip](https://ui.shadcn.com/charts/tooltip#charts) 的組合，提示顯示完整工作名稱和百分比。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
