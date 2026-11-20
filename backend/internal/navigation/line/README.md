# backend/internal/navigation/line 目錄導覽

## 用途與範圍

處理 LINE 導覽事件，依指令或互動引導使用者開啟對應選單與 MINI App 入口。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 server 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[events.go](events.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

從 [events.go](events.go)、[miniapp_test.go](miniapp_test.go)、[sticker_menu_test.go](sticker_menu_test.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
