# backend/internal/persistence/grpcstatus 目錄導覽

## 目錄用途

將 gRPC 的取消與逾時狀態對應到 Go context 錯誤，供資料存取層統一處理。

## 閱讀入口

從 [context.go](context.go)、[context_test.go](context_test.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
