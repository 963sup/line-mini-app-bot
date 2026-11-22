# Cloud Run Jobs 簡要規格

**狀態：按需導入。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

執行批次匯入、匯出、資料回填與統計重建。

授權者建立有範圍的執行請求 → Job 讀取批次 → 分段提交並記錄進度 → 回報成功／部分失敗。

## 資料、授權與失敗邊界

使用專用執行身分；批次按穩定 ID 及版本更新，保留重跑進度，不能覆寫較新的使用者資料。

## 導入條件

明定分片、併行上限、期限、重試、試跑、停止條件與恢復方式；大量資料不塞進互動 API。

## 驗收

中斷後續跑、重複執行、同時線上寫入、非法範圍及部分失敗可追查。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/run/docs/create-jobs)，查證範圍見[整合總覽](README.md#sources)。[資料演進](../../references/ddd/model-evolution.md) · [Scheduler](cloud-scheduler.md)。

[Google 服務索引](README.md)
