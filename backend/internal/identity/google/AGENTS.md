# backend/internal/identity/google 工作指引

## 目錄用途

驗證 Google 身分憑證，供後端確認呼叫端身分。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
