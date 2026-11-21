# backend/internal/transport/line/navigation 目錄導覽

管理 LINE 使用者的功能入口與選單，分開處理入口事件與回覆呈現。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [導覽事件](events.go)、[選單呈現](menu.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
