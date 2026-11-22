# Vertex AI／Google 生成式 AI 簡要規格

**狀態：按需導入；模型與端點於實作時選定。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

用於摘要、工程文件理解、分類或檢索輔助，不自行執行業務核准。

Go 核對可見資料 → 經模型 Adapter 提交限定內容 → 驗證輸出格式與來源 → 呈現建議或保存待確認結果。

## 資料、授權與失敗邊界

模型輸出及文件內指令視為不可信輸入；不得取得未授權內容、直接變更權限或跳過領域規則。

## 導入條件

指定模型／版本、區域、配額、成本、期限與資料政策；保留可替換的既有模型 Port，不預設強制遷移。

## 驗收

品質基準、來源追溯、格式錯誤、越權資料、提示注入及配額／服務失敗有預期結果。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/vertex-ai/docs/generative-ai/learn/overview)，查證範圍見[整合總覽](README.md#sources)。[Document AI](document-ai.md) · [業務操作模型](../../contracts/business-operations.md)。

[Google 服務索引](README.md)
