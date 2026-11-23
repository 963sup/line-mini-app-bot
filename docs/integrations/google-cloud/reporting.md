# 分析與報表

<a id="bigquery"></a>

## BigQuery

**狀態：按需導入。**

**用途與最小流程**

分析跨專案出勤、工程進度及長期趨勢，減少線上資料的分析掃描。

定義權威來源與指標 → 匯入有版本的分析資料 → 依分區查詢 → 對帳與發布報表。

**資料、授權與失敗邊界**

保留租戶／Repository 範圍與資料時間；分析結果不作即時寫入授權。資料集權限及去識別化依用途決定。

**導入條件**

指定 schema、分區、更新延遲、補跑去重、查詢成本與保存期；不直接將整庫個資複製到報表。

**驗收**

與來源對帳、遲到／更正資料、重跑、跨租戶隔離及查詢成本符合規格。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/bigquery/docs/introduction)。[報表](#looker-studio) · [資料演進](../../references/ddd/model-evolution.md)。

<a id="looker-studio"></a>

## Looker Studio／Data Studio 報表

**狀態：按需導入。**

**用途與最小流程**

呈現經授權的出勤、進度及趨勢管理報表。

選定受控資料集 → 定義指標、篩選與更新時間 → 設定資料憑證及分享範圍 → 驗收使用者可見內容。

**資料、授權與失敗邊界**

報表分享與資料讀權分開核對；不得假設 LINE Session 自動轉為 Google 報表權限。使用擁有者憑證時特別核對可被觀看的完整資料範圍。

**導入條件**

確定採用的報表產品及身分方案；資料列隔離須在資料來源／授權層落實，不能只靠 UI 篩選。

**驗收**

不同角色／租戶、匿名／外部分享、資料更新與撤權結果可驗證。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/data-studio/data-credentials)。[BigQuery](#bigquery) · [IAM](workload-identity.md#iam)。

[Google 服務索引與查證範圍](./)
