# backend/internal/business/journal 目錄導覽

## 目錄用途

處理工作日誌的查詢與寫入。`service.go` 維護業務流程，`http/` 對接 API，`firestore/` 保存及查詢日誌。

## 閱讀入口

從 [firestore](firestore)、[http](http)、[service.go](service.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
