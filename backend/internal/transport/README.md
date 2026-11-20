# backend/internal/transport 目錄導覽

## 用途與範圍

提供 HTTP 與 LINE 的共用傳輸工具，負責請求及回覆的協定處理。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 server 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[http](http)、[line](line)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

從 [http](http)、[line](line) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
