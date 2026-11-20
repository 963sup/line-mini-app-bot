# backend/internal/identity 工作指引

## 用途與範圍

定義共用身分及資源擁有者驗證，整合 LINE、Google 憑證驗證與應用程式工作階段。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[identity.go](identity.go)、[owner.go](owner.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：續期交易與前端觸發](../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：續期交易與前端觸發](../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../docs/diagrams.md#未來設計)
