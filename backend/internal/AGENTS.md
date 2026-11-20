# backend/internal 工作指引

## 目錄用途

保存 Go 後端的內部套件。業務功能放在 `business/`，服務組裝放在 `server/`，其餘套件提供身分驗證、請求傳輸、導覽及資料存取等共用基礎能力。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../docs/diagrams.md#未來設計)
