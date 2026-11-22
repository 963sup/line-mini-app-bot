# backend/internal/transport 目錄導覽

提供 HTTP 與 LINE 的共用傳輸工具，負責請求及回覆的協定處理。

[HTTP](http) 處理共用請求格式、靜態頁面與定位診斷。[LINE](line) 處理 Webhook 協定與事件派送，處理器由 bootstrap 依序註冊。

## 閱讀與修改入口

從 [http](http)、[line](line) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[上層導覽](../README.md)
