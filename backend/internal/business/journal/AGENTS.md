# backend/internal/business/journal 工作指引

## 閱讀與修改入口

先查閱[operations.go](application/port/operations.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 讀取已發布日誌的維護與驗收

發布後讀者失去來源權限時，仍能看到工作名稱與進度，但不能透過日誌取得來源名稱與識別碼。

修改此規則時，核對[判定流程](../../../../docs/domains/journal/README.md#decision-flow)與[代表測試](domain/entity/entry_test.go)。

[既有發布操作契約與流程](../../../../docs/domains/journal/journal.md#正式群組發布的操作契約)
