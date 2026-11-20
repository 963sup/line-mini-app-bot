# api/contract/journal 工作指引

## 用途與範圍

定義工作日誌的讀取與寫入契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[journals.yaml](journals.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口)

遵循[上層指引](../AGENTS.md)。調整公開請求或回應前，先核對契約與呼叫端，再執行 API 契約檢查。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../docs/diagrams.md#未來設計)
