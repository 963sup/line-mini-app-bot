# backend/internal/identity/http 目錄導覽

將身分驗證與工作階段接到 HTTP 層，處理登入請求及從 HTTP 請求取得身分。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [identity.go](identity.go)、[session.go](session.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發)

[上層導覽](../README.md)
