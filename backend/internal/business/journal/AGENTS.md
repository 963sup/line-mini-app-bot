# backend/internal/business/journal 工作指引

## 用途與範圍

處理工作日誌的查詢與寫入。`service.go` 維護業務流程，`http/` 對接 API，`firestore/` 保存及查詢日誌。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[service.go](application/port/operations.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 讀取已發布日誌的維護與驗收

發布後讀者失去來源權限時，仍能看到工作名稱與進度，但不能透過日誌取得來源名稱與識別碼。

修改此規則時，核對[判定流程](../../../../docs/domains/journal/README.md#decision-flow)與[代表測試](domain/entity/entry_test.go)。

[既有發布操作契約與流程](../../../../docs/domains/journal/journal.md#正式群組發布的操作契約)
