# backend/internal/business/summary 目錄導覽

## 用途與範圍

負責 LINE 群組訊息蒐集與摘要流程，涵蓋訊息保存、摘要產生、排程及交付。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[errors.go](errors.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：資料與可靠性](../../../../docs/domains/summary.md#資料與可靠性)

## 閱讀入口

從 [errors.go](errors.go)、[firestore](firestore)、[line](line)、[messagecollection](messagecollection)、[summarization](summarization) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：資料與可靠性](../../../../docs/domains/summary.md#資料與可靠性) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
