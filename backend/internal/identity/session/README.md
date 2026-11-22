# backend/internal/identity/session 目錄導覽

管理應用程式工作階段的建立、驗證與有效期限，`firestore/` 負責保存工作階段。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [session.go](session.go)、[firestore](firestore) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：續期交易與前端觸發](../../../../docs/contracts/identity.md#續期交易與前端觸發)

[上層導覽](../README.md)
