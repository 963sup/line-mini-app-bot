# backend/internal/server 目錄導覽

## 目錄用途

組裝後端服務、業務依賴與各功能路由，將 HTTP、LINE 與資料存取轉接器接到業務層。

## 閱讀入口

從 [miniapp.go](miniapp.go)、[miniapp_test.go](miniapp_test.go)、[attendance.go](attendance.go)、[group.go](group.go)、[journal.go](journal.go)、[line.go](line.go)、[my_engineering.go](my_engineering.go)、[my_engineering_test.go](my_engineering_test.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
