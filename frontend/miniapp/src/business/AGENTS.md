# frontend/miniapp/src/business 工作指引

## 架構與責任邊界

顯示內容編輯控制時使用後端回傳的 `canEdit`。測試應包含關閉該能力時的畫面，並保留具備獨立授權的分類操作。切換帳號或群組後，重新取得該範圍的資料。

## 閱讀與修改入口

先查閱[attendance](attendance)、[discussion](discussion)、[enterprise](enterprise)、[issue](issue)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
