# backend/internal/transport/http/browser 目錄導覽

提供瀏覽器頁面路由與靜態檔案服務，負責將頁面請求交給對應資源。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [files.go](files.go)、[routes.go](routes.go)、[routes_test.go](routes_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

`/miniapp/poppler` 提供 PDF 文字擷取工具。建置檢查包含該頁面，避免只產出靜態檔而未接入 Go 路由。`/_next/static/chunks/` 的 JS／CSS 使用一年 immutable 快取；這些檔名由 Next 建置產生版本雜湊。HTML、其他資源及私人 API 維持原有不快取政策。
