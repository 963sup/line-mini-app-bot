# backend/internal/identity/line 目錄導覽

驗證 LINE 身分憑證，供 MINI App 登入建立可信的使用者身分。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [verify.go](verify.go)、[verify_test.go](verify_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：續期交易與前端觸發](../../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
