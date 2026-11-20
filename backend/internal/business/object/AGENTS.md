# backend/internal/business/object 工作指引

## 用途與範圍

管理工程清單物件的類型、精確數量、單位與內部追蹤狀態。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[object.go](domain/entity/object.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 變更工項進度的維護與驗收

總量為 5、完成量為 8 的工項應被拒絕；等值十進位表示不應產生不同的完成狀態。

修改此規則時，核對[判定流程](../../../../docs/domains/object/README.md#decision-flow)與[代表測試](application/progress_test.go)。
