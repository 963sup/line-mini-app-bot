# backend/internal/transport/http 目錄導覽

提供共用 HTTP JSON 讀寫工具。[api](api) 接線 OpenAPI 契約，[browser](browser) 供應靜態頁面，[geolocation](geolocation) 處理定位診斷。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [json.go](json.go)、[read_json_test.go](read_json_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[上層導覽](../README.md)
