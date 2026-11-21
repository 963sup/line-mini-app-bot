# backend/internal/platform 工作指引

## 用途與範圍

提供外部平台與資料存取的共用機制，包含 LINE 群組成員查詢、Firestore 鍵值與游標處理，以及 gRPC 錯誤對應。

## 架構與責任邊界

共用的鍵值編碼、游標與錯誤處理可由多個適配器使用。業務文件映射、交易與授權規則由對應的業務模組維護。

## 閱讀與修改入口

先查閱[line](line/README.md)、[firestorekey](firestorekey)、[grpcstatus](grpcstatus)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
