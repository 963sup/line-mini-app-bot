# backend/internal/transport/line 目錄導覽

## 目錄用途

提供 LINE Webhook 簽章驗證與事件處理的傳輸工具，以及回覆和群組成員查詢的共用介面。

## 閱讀入口

從 [group.go](group.go)、[group_test.go](group_test.go)、[membership.go](membership.go)、[membership_test.go](membership_test.go)、[reply.go](reply.go)、[webhook.go](webhook.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：基礎入口](../../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
