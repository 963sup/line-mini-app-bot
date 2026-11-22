# backend/internal/business/journal 目錄導覽

[領域文件](../../../../docs/domains/journal/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

處理工作日誌的查詢與寫入。`service.go` 維護業務流程，`http/` 對接 API，`firestore/` 保存及查詢日誌。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

從 [operations.go](application/port/operations.go)、[firestore](adapter/firestore)、[http](adapter/http) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 領域追蹤入口

[領域規格](../../../../docs/domains/journal/README.md) · [判定流程](../../../../docs/domains/journal/README.md#decision-flow) · [代表測試](domain/entity/entry_test.go)

[既有發布操作契約與流程](../../../../docs/domains/journal/journal.md#正式群組發布的操作契約)

[上層導覽](../README.md)
