# 文件工具契約

`poppler.yaml` 定義登入使用者的 PDF 文字擷取測試。JSON 內含檔名與 Base64 資料，原始 PDF 上限 8 MiB；回傳標準化文字，保留分頁 FF。操作不儲存文件、不呼叫 AI，也不建立業務物件。Cookie Session 需要 CSRF token，並套用同源保護。

來源組合於 `../openapi.yaml`，生成與驗證依上層 API 指引執行 `pnpm api:generate`、`pnpm api:check`。
