# backend/internal/httpapi 目錄導覽

## 目錄用途

將 OpenAPI 契約接入 Go HTTP 路由，並保存公開與內部 API 的產生碼。契約變更從根目錄的 `api/` 開始。

## 閱讀入口

從 [contract.go](contract.go)、[generated](generated)、[generatedinternal](generatedinternal)、[routes.go](routes.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
