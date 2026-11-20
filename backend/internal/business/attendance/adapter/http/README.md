# Attendance http

HTTP 請求、認證接線與契約轉換。

## 實作與規則

處理 HTTP 與 OpenAPI DTO 映射，入口為 [routes.go](routes.go)。身分來源是伺服器驗證結果，前端不能決定審核權。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
