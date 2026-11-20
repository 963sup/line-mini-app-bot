# backend/internal/identity/session 目錄導覽

## 目錄用途

管理應用程式工作階段的建立、驗證與有效期限，`firestore/` 負責保存工作階段。

## 閱讀入口

從 [firestore](firestore)、[session.go](session.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
