# backend/internal/platform/grpcstatus 目錄導覽

將 gRPC 的取消與逾時狀態對應到 Go context 錯誤，供資料存取層統一處理。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [context.go](context.go)、[context_test.go](context_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
