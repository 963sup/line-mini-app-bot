# api/contract/object 目錄導覽

## 用途與範圍

定義工程清單物件與子物件的 API 契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[objects.yaml](objects.yaml)、[subobjects.yaml](subobjects.yaml)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：物件與子物件歸屬](../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬)

物件操作見 [objects.yaml](objects.yaml)，子物件操作見 [subobjects.yaml](subobjects.yaml)。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：物件與子物件歸屬](../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

[ObjectTemplate 最小閉環](../../../docs/domains/object/template.md)：建立不可變 v1、選用、原子套用並開啟獨立 Object 樹；草稿、改版與舊資料遷移留在後續範圍。
