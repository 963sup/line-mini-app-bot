# backend/internal/identity 目錄導覽

## 用途與範圍

定義共用身分及資源擁有者驗證，整合 LINE、Google 憑證驗證與應用程式工作階段。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[identity.go](identity.go)、[owner.go](owner.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：續期交易與前端觸發](../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發)

## 閱讀入口

從 [google](google)、[group_owner_test.go](group_owner_test.go)、[http](http)、[identity.go](identity.go)、[line](line)、[owner.go](owner.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：續期交易與前端觸發](../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

## 下層文件入口

- [session](session/README.md)
