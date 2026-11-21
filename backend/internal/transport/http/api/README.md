# backend/internal/transport/http/api 目錄導覽

將 OpenAPI 契約接入 Go HTTP 路由，並保存公開與內部 API 的產生碼。契約變更從根目錄的 `api/` 開始。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [contract.go](contract.go)、[routes.go](routes.go)、[generated](generated)、[generatedinternal](generatedinternal) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
