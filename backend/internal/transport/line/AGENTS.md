# backend/internal/transport/line 工作指引

## 目錄用途

提供 LINE Webhook 簽章驗證與事件處理的傳輸工具，以及回覆和群組成員查詢的共用介面。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：基礎入口](../../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
