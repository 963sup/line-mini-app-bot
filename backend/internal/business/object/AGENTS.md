# backend/internal/business/object 工作指引

## 閱讀與修改入口

先查閱[object.go](domain/entity/object.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 變更工項進度的維護與驗收

總量為 5、完成量為 8 的工項應被拒絕；等值十進位表示不應產生不同的完成狀態。

修改此規則時，核對[判定流程](../../../../docs/domains/object/README.md#decision-flow)與[代表測試](application/progress_test.go)。
