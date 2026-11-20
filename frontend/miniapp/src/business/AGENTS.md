# frontend/miniapp/src/business 工作指引

## 目錄用途

遵循 **依功能組織（feature-based architecture）**。將同一功能的畫面、表單、狀態與型別化 API 呼叫就近維護，由 `pages/` 依路由組合。前端呈現輸入與操作結果，業務規則及權限由後端做最終判定。

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
