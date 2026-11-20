# backend/internal/business/journal 工作指引

## 目錄用途

處理工作日誌的查詢與寫入。`service.go` 維護業務流程，`http/` 對接 API，`firestore/` 保存及查詢日誌。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
