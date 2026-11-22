# 空間問題

[領域文件](../../../../docs/domains/repository/issue/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

管理 Repository 中的問題事項及其狀態，維護問題專屬的操作與授權規則。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[issue.go](domain/entity/issue.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[問題規格](../../../../docs/domains/repository/issue/design.md) · [現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

操作權限、狀態、API、資料與驗收依 [問題規格](../../../../docs/domains/repository/issue/design.md) 維護；本模組只保存對應業務實作。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 領域追蹤入口

[領域規格](../../../../docs/domains/repository/issue/README.md) · [判定流程](../../../../docs/domains/repository/issue/README.md#decision-flow) · [代表測試](domain/entity/issue_test.go)

[上層導覽](../README.md)
