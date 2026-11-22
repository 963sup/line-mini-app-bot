# Looker Studio／Data Studio 報表 簡要規格

**狀態：按需導入。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

呈現經授權的出勤、進度及趨勢管理報表。

選定受控資料集 → 定義指標、篩選與更新時間 → 設定資料憑證及分享範圍 → 驗收使用者可見內容。

## 資料、授權與失敗邊界

報表分享與資料讀權分開核對；不得假設 LINE Session 自動轉為 Google 報表權限。使用擁有者憑證時特別核對可被觀看的完整資料範圍。

## 導入條件

確定採用的報表產品及身分方案；資料列隔離須在資料來源／授權層落實，不能只靠 UI 篩選。

## 驗收

不同角色／租戶、匿名／外部分享、資料更新與撤權結果可驗證。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/data-studio/data-credentials)，查證範圍見[整合總覽](README.md#sources)。[BigQuery](bigquery.md) · [IAM](iam.md)。

[Google 服務索引](README.md)
