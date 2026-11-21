# api/client/src 目錄導覽

提供 client 建立與請求憑證設定。`generated/` 的型別由公開 OpenAPI 契約產生，欄位變更應回到契約來源。

契約是介面來源；client 與產生設定供消費端使用，依契約重新產生型別。

## 閱讀與修改入口

從 [client.ts](client.ts)、[generated](generated) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
