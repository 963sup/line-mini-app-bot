# backend/internal/identity/http 目錄導覽

## 用途與範圍

將身分驗證與工作階段接到 HTTP 層，處理登入請求及從 HTTP 請求取得身分。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 server 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[identity.go](identity.go)、[session.go](session.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發)

## 閱讀入口

從 [identity.go](identity.go)、[session.go](session.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
