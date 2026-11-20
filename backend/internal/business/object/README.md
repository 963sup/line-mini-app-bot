# 儲存庫物件

[領域文件](../../../../docs/domains/object/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

## 用途與範圍

管理工程清單物件的類型、精確數量、單位與內部追蹤狀態。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[object.go](domain/entity/object.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[物件規格](../../../../docs/domains/object/design.md) · [現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## 閱讀入口

Object 保存工程清單項的類型、精確數量、單位及內部追蹤狀態，與 Issue 問題分工。
操作權限、來源及驗收依 [物件規格](../../../../docs/domains/object/design.md) 維護。

[現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 領域追蹤入口

[領域規格](../../../../docs/domains/object/README.md) · [判定流程](../../../../docs/domains/object/README.md#decision-flow) · [代表測試](application/progress_test.go)

[上層導覽](../README.md)
