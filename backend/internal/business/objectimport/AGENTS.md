# backend/internal/business/objectimport 工作指引

## 用途與範圍

處理工程清單匯入，從來源解析、候選驗證和人工審核，到確認後透過 Object 使用案例建立物件。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[object_import.go](domain/entity/object_import.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：自動匯入操作契約](../../../../docs/domains/objectimport/design.md#自動匯入操作契約)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：自動匯入操作契約](../../../../docs/domains/objectimport/design.md#自動匯入操作契約) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 查閱自動匯入紀錄的維護與驗收

以其他使用者的身分操作已有的自動匯入紀錄時，回傳找不到，避免揭露該筆紀錄是否存在。

修改此規則時，核對[判定流程](../../../../docs/domains/objectimport/README.md#decision-flow)與[代表測試](adapter/firestore/automatic_test.go)。
