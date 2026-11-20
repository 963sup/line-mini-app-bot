# api/client/src 工作指引

## 目錄用途

提供 client 建立與請求憑證設定。`generated/` 的型別由公開 OpenAPI 契約產生，欄位變更應回到契約來源。

遵循[上層指引](../AGENTS.md)。調整公開請求或回應前，先核對契約與呼叫端，再執行 API 契約檢查。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../docs/diagrams.md#未來設計)
