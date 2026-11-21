# api/contract/object 工作指引

## 用途與範圍

定義工程清單物件與子物件的 API 契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[objects.yaml](objects.yaml)、[subobjects.yaml](subobjects.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：物件與子物件歸屬](../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬)

遵循[上層指引](../AGENTS.md)。調整公開請求或回應前，先核對契約與呼叫端，再執行 API 契約檢查。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：物件與子物件歸屬](../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../docs/diagrams.md#未來設計)
