# backend/internal/platform 目錄導覽

提供外部平台與資料存取的共用機制，包含 LINE 群組成員查詢、Firestore 鍵值與游標處理，以及 gRPC 錯誤對應。

共用的鍵值編碼、游標與錯誤處理可由多個適配器使用。業務文件映射、交易與授權規則由對應的業務模組維護。

## 閱讀與修改入口

從 [line](line/README.md)、[firestorekey](firestorekey)、[grpcstatus](grpcstatus) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
