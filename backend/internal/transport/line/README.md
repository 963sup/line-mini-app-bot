# backend/internal/transport/line 目錄導覽

## 用途與範圍

提供 LINE Webhook 簽章驗證與事件處理的傳輸工具，以及回覆協定。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

[navigation](navigation) 處理選單指令、postback 與回覆呈現。業務專屬指令由各領域的 LINE 適配器處理。

## 閱讀與修改入口

先查閱[group.go](group.go)、[reply.go](reply.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../../../docs/architecture/contracts/http.md#基礎入口)

## 閱讀入口

從 [group.go](group.go)、[group_test.go](group_test.go)、[reply.go](reply.go)、[webhook.go](webhook.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：基礎入口](../../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

群組成員 API 查詢由 [platform/line](../../platform/line/README.md) 實作，透過 bootstrap 注入各領域的 Membership port。
