# backend/internal/business 目錄導覽

## 目錄用途

採用 **依業務領域拆分（package by domain）**。業務規則與使用案例放在負責該操作的套件，HTTP、LINE 與 Firestore 轉接器依領域分工。跨領域操作透過明確的介面協作；服務與路由的組裝由 `internal/server` 負責。

## 閱讀入口

從 [attendance](attendance)、[discussion](discussion)、[group](group)、[issue](issue)、[journal](journal)、[object](object) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
