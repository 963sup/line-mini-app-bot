# 四領域協作提案

[Workflow](workflow.md) 管理步驟順序與等待條件，透過 [Operations](operations.md) 呼叫業務操作。資源所屬領域負責驗證授權、資料版本與寫入規則。

需要補查證據時，[Agent](agent.md) 可在授權範圍內選擇操作。[Knowledge](knowledge.md) 提供已發布且適用於目前組織的規則版本。

## 物件呼叫文件辨識能力

Object 的應用層先核對呼叫者與儲存庫，再呼叫 Operations 的文件候選擷取介面。Bootstrap 按 `AI_DOCUMENT_MODEL` 和既有 OpenAI 憑證注入適配器。這條內部接線只回傳候選；正式物件仍由 Object 建立。

適配器透過 Responses API 在同一個請求中使用 `input_file` 傳送原檔，並以 `text.format` 指定嚴格的 JSON Schema。候選包含名稱、十進位數量字串、單位、來源位置與引文；數量、單位及位置允許 null。完成、拒答、截斷和無效輸出各有對應的結果，呼叫端可檢視模型回應。每次重送都會重新辨識且可能計費。[File inputs](https://developers.openai.com/api/docs/guides/file-inputs?api-mode=responses) · [Structured Outputs](https://developers.openai.com/api/docs/guides/structured-outputs#supported-schemas)

### 後續實作順序

1. 在物件 API 加入候選辨識操作，定義檔案類型、上傳大小、候選及錯誤回應的公開契約。測試無權限時不會呼叫模型，拒答與截斷可和零筆候選區分。
2. 在物件畫面加入選檔與覆核。顯示原始數量、單位、位置與引文，讓使用者修正後透過既有建立操作逐筆送出。重送建立時沿用同一建立鍵。
3. 以實際 PDF、段落文件及試算表驗收來源可追溯性、缺欄與漏列。試算表處理有列數及摘要限制，需用原件核對完整性。

跨步驟續跑、補查外部證據及套用已發布規則分別由 Workflow、Agent 和 Knowledge 按需求承接。

## 驗收條件

- 未授權的呼叫無法讀取資源或執行寫入。
- 重試同一次寫入只產生一次業務結果。
- 補查只使用允許的來源；引用的知識版本可追溯。
