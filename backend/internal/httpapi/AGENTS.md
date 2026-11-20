# backend/internal/httpapi 工作指引

## 目錄用途

將 OpenAPI 契約接入 Go HTTP 路由，並保存公開與內部 API 的產生碼。契約變更從根目錄的 `api/` 開始。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
