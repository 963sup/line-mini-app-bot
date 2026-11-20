# backend/internal/identity/http 目錄導覽

## 目錄用途

將身分驗證與工作階段接到 HTTP 層，處理登入請求及從 HTTP 請求取得身分。

## 閱讀入口

從 [identity.go](identity.go)、[session.go](session.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
