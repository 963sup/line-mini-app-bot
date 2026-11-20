# backend/internal/business 工作指引

## 用途與範圍

採用 **依業務領域拆分（package by domain）**。業務規則與使用案例放在負責該操作的套件，HTTP、LINE 與 Firestore 轉接器依領域分工。跨領域操作透過明確的介面協作；服務與路由的組裝由 `internal/bootstrap` 負責。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[attendance](attendance)、[discussion](discussion)、[enterprise](enterprise)、[group](group)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
