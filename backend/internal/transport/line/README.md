# backend/internal/transport/line 目錄導覽

## 用途與範圍

提供 LINE Webhook 簽章驗證與事件處理的傳輸工具，以及回覆和群組成員查詢的共用介面。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 server 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[group.go](group.go)、[membership.go](membership.go)、[reply.go](reply.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../../../docs/architecture/contracts/http.md#基礎入口)

## 閱讀入口

從 [group.go](group.go)、[group_test.go](group_test.go)、[membership.go](membership.go)、[membership_test.go](membership_test.go)、[reply.go](reply.go)、[webhook.go](webhook.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：基礎入口](../../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
