# backend/cmd/firestorecheck 工作指引

## 用途與範圍

檢查 Firestore 連線與資料存取的 Go 指令入口。

## 架構與責任邊界

指令負責解析輸入、組裝依賴與執行指定工作；共用服務與業務規則由 internal 套件維護。

## 閱讀與修改入口

先查閱[main.go](main.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)


遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
