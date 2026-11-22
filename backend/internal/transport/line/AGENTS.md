# backend/internal/transport/line 工作指引

## 閱讀與修改入口

先查閱[group.go](group.go)、[reply.go](reply.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 工作方式

[現行實作圖：基礎入口](../../../../docs/contracts/http.md#基礎入口)

群組成員 API 查詢由 [platform/line](../../platform/line/README.md) 實作，透過 bootstrap 注入各領域的 Membership port。
