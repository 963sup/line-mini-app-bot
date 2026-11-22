# backend/internal/business 目錄導覽

採用 **依業務領域拆分（package by domain）**。業務規則與使用案例放在負責該操作的套件，HTTP、LINE 與 Firestore 轉接器依領域分工。跨領域操作透過明確的介面協作；服務與路由的組裝由 `internal/bootstrap` 負責。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

從 [attendance](attendance)、[discussion](discussion)、[enterprise](enterprise)、[issue](issue)、[journal](journal)、[object](object) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 規劃中的業務領域

[operations](operations/README.md) · [workflow](workflow/README.md) · [agent](agent/README.md) · [knowledge](knowledge/README.md)

[上層導覽](../README.md)
