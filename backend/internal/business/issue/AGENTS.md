# backend/internal/business/issue 工作指引

## 用途與範圍

管理 Repository 中的問題事項及其狀態，維護問題專屬的操作與授權規則。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[issue.go](domain/entity/issue.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 更新議題的維護與驗收

建立者可關閉議題後重新開啟；只有分類權限的成員不能藉由分類操作修改內容。

修改此規則時，核對[判定流程](../../../../docs/domains/issue/README.md#decision-flow)與[代表測試](domain/entity/issue_test.go)。
