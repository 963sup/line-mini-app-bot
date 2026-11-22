# backend/internal/bootstrap 目錄導覽

組裝後端服務、業務依賴與各功能路由，將 HTTP、LINE 與資料存取轉接器接到業務層。

啟動入口在 [server.go](server.go)，負責設定、路由註冊、資源釋放及關機。修改 Webhook 處理順序時查閱 [line.go](line.go)；組織建立對話與工程清單回覆由各自的業務 LINE 適配器處理。

## 閱讀與修改入口

從 [attendance.go](attendance.go)、[creation_line.go](../business/organization/adapter/line/creation_line.go)、[group.go](group.go)、[miniapp.go](miniapp.go)、[miniapp_test.go](miniapp_test.go)、[journal.go](journal.go)、[line.go](line.go)、[my_engineering.go](../business/repository/adapter/line/my_engineering.go)、[my_engineering_test.go](../business/repository/adapter/line/my_engineering_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[上層導覽](../README.md)

部署 Job 的 LINE 選單同步由 [line_navigation.go](line_navigation.go) 組裝，使用 Firestore 租約與有界 context；HTTP 服務啟動不發布。租約整合測試需隔離的 loopback Firestore emulator。
