# backend/internal/transport 工作指引

## 用途與範圍

提供 HTTP 與 LINE 的共用傳輸工具，負責請求及回覆的協定處理。

## 架構與責任邊界

[HTTP](http) 處理共用請求格式、靜態頁面與定位診斷。[LINE](line) 處理 Webhook 協定與事件派送，處理器由 bootstrap 依序註冊。

## 閱讀與修改入口

先查閱[http](http)、[line](line)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
