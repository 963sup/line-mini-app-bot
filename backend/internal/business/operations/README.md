# operations

[領域文件](../../../../docs/domains/operations/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

提供可重用的具型別操作，將請求委派給擁有業務規則的領域。

[操作與授權流程圖](../../../../docs/proposals/operations.md#design-diagram) · [維護指引](AGENTS.md)

[上層導覽](../README.md)

## PDF 文字擷取基礎接線

`application.DocumentTextExtractor` 定義文字擷取介面，`adapter/poppler` 使用執行環境的 `pdftotext` 實作。bootstrap 將 adapter 注入 object service，預留給未來匯入用例；工具測試用例透過獨立的 `DocumentTextService` 呼叫此介面，既有 OpenAI 文件解析及 Firestore 寫入流程維持原行為。Dockerfile 的執行階段安裝 `poppler-utils`。

Adapter 使用 stdin/stdout，不建立暫存檔；限制輸入及文字輸出各 8 MiB、轉換時間 30 秒，保留 UTF-8 文字、版面與分頁。尚未實作文字轉業務 JSON 或自動儲存。

Poppler 輸出固定為 `-layout -enc UTF-8 -eol unix`。adapter 驗證 UTF-8，統一 CRLF／CR 為 LF，移除 C0／C1 控制字元但保留 LF、Tab 與分頁 FF；不壓縮空白、不做 Unicode NFC。空文字仍是有效擷取結果，不能據此斷言文件需要 OCR。

錯誤可用 `errors.Is` 辨識工具缺失、輸出超限、PDF 開啟失敗（退出碼 1）、輸出開啟失敗（2）、PDF 權限問題（3）、其餘轉換錯誤（含 99）、無效 UTF-8，以及 context 取消／逾時；同時保留 application 的文件或 provider 錯誤分類。選項與退出碼依 Poppler 官方 `utils/pdftotext.1`，經 Context7 匿名查詢 master 核對；本機版本為 25.03.0，映像套件由 Debian Bookworm 提供，部署版本仍需於映像驗收。

## PDF 工具測試

側欄「工具 → PDF 文字擷取」開啟 `/miniapp/poppler`。登入後可上傳 8 MiB 以內的 PDF，透過 `POST /api/v1/tools/poppler/extractions` 取得標準化文字，畫面顯示分頁標記並可下載保留 FF 的文字原檔。空文字顯示提示。工具只處理單次請求，不保存文件或文字，不呼叫 OpenAI 或 Firestore；可重試且無持久副作用。

HTTP adapter 使用既有身分驗證、Cookie CSRF 及同源保護，限制 JSON request body 為 12 MiB；application 驗證可信 Actor，Poppler adapter 執行轉換。失敗回傳安全的錯誤代碼與說明，不回傳程序 stderr 或文件內容。
