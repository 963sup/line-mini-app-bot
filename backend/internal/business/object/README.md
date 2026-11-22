# 儲存庫物件

[領域文件](../../../../docs/domains/repository/object/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

管理工程清單物件的類型、精確數量、單位與內部追蹤狀態。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[object.go](domain/entity/object.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[物件規格](../../../../docs/domains/repository/object/design.md) · [現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## 閱讀入口

Object 保存工程清單項的類型、精確數量、單位及內部追蹤狀態，與 Issue 問題分工。
操作權限、來源及驗收依 [物件規格](../../../../docs/domains/repository/object/design.md) 維護。

[現行實作圖：物件與子物件歸屬](../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 領域追蹤入口

[領域規格](../../../../docs/domains/repository/object/README.md) · [判定流程](../../../../docs/domains/repository/object/README.md#decision-flow) · [代表測試](application/progress_test.go)

[上層導覽](../README.md)

物件可只含名稱，類型、數量、單位與空間允許未設定。未設定不是數量0或預設施作；已有完成量的施作仍受精確數量、單位與進度移除規則保護。Firestore 沿用原有欄位與 schemaVersion，既有資料保持可讀。

[ObjectTemplate 最小閉環](../../../../docs/domains/repository/object/template.md)：建立不可變 v1、選用、原子套用並開啟獨立 Object 樹；草稿、改版與舊資料遷移留在後續範圍。
