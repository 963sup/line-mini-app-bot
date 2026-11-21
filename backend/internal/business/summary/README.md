# backend/internal/business/summary 目錄導覽

[領域文件](../../../../docs/domains/summary/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

負責 LINE 群組訊息蒐集與摘要流程，涵蓋訊息保存、摘要產生、排程及交付。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

從 [errors.go](application/port/errors.go)、[firestore](adapter/firestore)、[line](adapter/line)、[messagecollection](application/messagecollection)、[summarization](application/summarization) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：資料與可靠性](../../../../docs/domains/summary/summary.md#資料與可靠性) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 領域追蹤入口

[領域規格](../../../../docs/domains/summary/README.md) · [判定流程](../../../../docs/domains/summary/README.md#decision-flow) · [判定實作](application/summarization/validation.go)

[上層導覽](../README.md)
