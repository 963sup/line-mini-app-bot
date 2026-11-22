# backend/internal/platform/firestore/key 目錄導覽

提供 Firestore 文件鍵值與分頁游標的共用處理。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [cursor.go](cursor.go)、[key.go](key.go)、[cursor_test.go](cursor_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[上層導覽](../../README.md)
