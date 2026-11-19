# frontend/miniapp/src/business/journal 目錄導覽

從 [Workspace.tsx](Workspace.tsx) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

- Feed：按「寫日誌」→ 選擇發布群組 → 填寫並發布。工作牆的群組篩選與發布群組分開。
- Dashboard：依日期及群組篩選，透過 [GroupCharts.tsx](GroupCharts.tsx) 為每個有日誌的群組顯示一張 shadcn 長條圖，逐項呈現今日實際及明日預計累計完成度。載入更多時，同群組的日誌會加到同一張圖表。展開日誌明細可查看人員及原文連結。
- Dashboard 圖表採用 shadcn 官方 [Bar](https://ui.shadcn.com/charts/bar#charts) 與 [Tooltip](https://ui.shadcn.com/charts/tooltip#charts) 的組合，提示顯示完整工作名稱和百分比。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
