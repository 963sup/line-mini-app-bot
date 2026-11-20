# frontend/miniapp/src/business 工作指引

## 用途與範圍

遵循 **依功能組織（feature-based architecture）**。將同一功能的畫面、表單、狀態與型別化 API 呼叫就近維護，由 `pages/` 依路由組合。前端呈現輸入與操作結果，業務規則及權限由後端做最終判定。

## 架構與責任邊界

顯示內容編輯控制時使用後端回傳的 `canEdit`。測試應包含關閉該能力時的畫面，並保留具備獨立授權的分類操作。切換帳號或群組後，重新取得該範圍的資料。

## 閱讀與修改入口

先查閱[attendance](attendance)、[discussion](discussion)、[enterprise](enterprise)、[issue](issue)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
