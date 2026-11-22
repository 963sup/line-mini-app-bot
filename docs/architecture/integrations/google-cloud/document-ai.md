# Document AI 簡要規格

**狀態：按需導入。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

從工程 PDF、掃描文件及表格擷取文字與候選欄位。

授權取得檔案版本 → 背景送交指定 processor → 保存來源頁碼與擷取結果 → 人工確認後建立業務資料。

## 資料、授權與失敗邊界

原檔由 Storage 保存；只處理有權文件，記錄 processor／版本、來源 generation 與錯誤，不把 OCR 結果直接當已核准數量。

## 導入條件

確認格式、語言、區域、頁量與成本；低信心、缺頁、欄位錯配進待確認流程。

## 驗收

代表文件的辨識率與人工修正可驗證；同版本重跑、非法文件及處理失敗不重複建立工項。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/document-ai/docs/overview)，查證範圍見[整合總覽](README.md#sources)。[Cloud Storage](../../../integrations/storage/cloud-storage.md) · [Tasks](cloud-tasks.md)。

[Google 服務索引](README.md)
