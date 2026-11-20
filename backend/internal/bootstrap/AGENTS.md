# backend/internal/bootstrap 工作指引

## 用途與範圍

組裝後端服務、業務依賴與各功能路由，將 HTTP、LINE 與資料存取轉接器接到業務層。

## 架構與責任邊界

啟動入口在 [server.go](server.go)，負責設定、路由註冊、資源釋放及關機。修改 Webhook 處理順序時查閱 [line.go](line.go)；組織建立對話與工程清單回覆由各自的業務 LINE 適配器處理。

## 閱讀與修改入口

先查閱[attendance.go](attendance.go)、[creation_line.go](../business/organization/adapter/line/creation_line.go)、[group.go](group.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
