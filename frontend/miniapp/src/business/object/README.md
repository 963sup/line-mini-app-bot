# 儲存庫物件

## 用途與範圍

提供工程物件清單、詳情與編輯畫面。

## 架構與責任邊界

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[View.tsx](View.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[物件規格](../../../../../docs/domains/object/design.md)

提供物件清單、詳情、手動新增及建立者編輯；欄位、權限與狀態依 [物件規格](../../../../../docs/domains/object/design.md)。

`View` 依儲存庫及身分世代重設狀態，未知建立結果保留原請求鍵及內容；父層導覽在建立結果未確認時保持目前表單。沿用既有 typed client、請求取消及官方 UI 元件組合。

[現行實作圖：物件與子物件歸屬](../../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

新增物件僅顯示名稱，不載入空間選項；其他資料確定後再從詳情編輯補充。未分類物件不顯示進度與狀態。來源候選覆核保留原文證據，建立時不自動套入未確認的數量或分類。

[ObjectTemplate 最小閉環](../../../../../docs/domains/object/template.md)：建立不可變 v1、選用、原子套用並開啟獨立 Object 樹；草稿、改版與舊資料遷移留在後續範圍。
