# backend/internal/bootstrap 目錄導覽

## 用途與範圍

組裝後端服務、業務依賴與各功能路由，將 HTTP、LINE 與資料存取轉接器接到業務層。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[attendance.go](attendance.go)、[creation_line.go](../business/organization/adapter/line/creation_line.go)、[group.go](group.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

從 [miniapp.go](miniapp.go)、[miniapp_test.go](miniapp_test.go)、[attendance.go](attendance.go)、[group.go](group.go)、[journal.go](journal.go)、[line.go](line.go)、[my_engineering.go](../business/repository/adapter/line/my_engineering.go)、[my_engineering_test.go](../business/repository/adapter/line/my_engineering_test.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
