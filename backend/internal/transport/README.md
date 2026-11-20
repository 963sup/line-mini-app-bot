# backend/internal/transport 目錄導覽

## 用途與範圍

提供 HTTP 與 LINE 的共用傳輸工具，負責請求及回覆的協定處理。

## 架構與責任邊界

[HTTP](http) 處理共用請求格式、靜態頁面與定位診斷。[LINE](line) 處理 Webhook 協定與事件派送，處理器由 bootstrap 依序註冊。

## 閱讀與修改入口

先查閱[http](http)、[line](line)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

從 [http](http)、[line](line) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
