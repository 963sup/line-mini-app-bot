# backend/internal/business/summary 工作指引

## 用途與範圍

負責 LINE 群組訊息蒐集與摘要流程，涵蓋訊息保存、摘要產生、排程及交付。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[errors.go](application/port/errors.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：資料與可靠性](../../../../docs/domains/summary/summary.md#資料與可靠性)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：資料與可靠性](../../../../docs/domains/summary/summary.md#資料與可靠性) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
