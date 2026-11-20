# backend/internal/identity/session 工作指引

## 目錄用途

管理應用程式工作階段的建立、驗證與有效期限，`firestore/` 負責保存工作階段。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
