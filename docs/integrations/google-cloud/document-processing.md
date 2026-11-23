# 文件擷取與模型推論

<a id="document-ai"></a>

## Document AI

**狀態：按需導入。**

**用途與最小流程**

從工程 PDF、掃描文件及表格擷取文字與候選欄位。

授權取得檔案版本 → 背景送交指定 processor → 保存來源頁碼與擷取結果 → 人工確認後建立業務資料。

**資料、授權與失敗邊界**

原檔由 Storage 保存；只處理有權文件，記錄 processor／版本、來源 generation 與錯誤，不把 OCR 結果直接當已核准數量。

**導入條件**

確認格式、語言、區域、頁量與成本；低信心、缺頁、欄位錯配進待確認流程。

**驗收**

代表文件的辨識率與人工修正可驗證；同版本重跑、非法文件及處理失敗不重複建立工項。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/document-ai/docs/overview)。[Cloud Storage](../storage/cloud-storage.md) · [Tasks](background-jobs.md#cloud-tasks)。

<a id="vertex-ai"></a>

## Vertex AI／Google 生成式 AI

**狀態：按需導入；模型與端點於實作時選定。**

**用途與最小流程**

用於摘要、工程文件理解、分類或檢索輔助，不自行執行業務核准。

Go 核對可見資料 → 經模型 Adapter 提交限定內容 → 驗證輸出格式與來源 → 呈現建議或保存待確認結果。

**資料、授權與失敗邊界**

模型輸出及文件內指令視為不可信輸入；不得取得未授權內容、直接變更權限或跳過領域規則。

**導入條件**

指定模型／版本、區域、配額、成本、期限與資料政策；保留可替換的既有模型 Port，不預設強制遷移。

**驗收**

品質基準、來源追溯、格式錯誤、越權資料、提示注入及配額／服務失敗有預期結果。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/vertex-ai/docs/generative-ai/learn/overview)。[Document AI](#document-ai) · [業務操作模型](../../contracts/business-operations.md)。

[Google 服務索引與查證範圍](./)
