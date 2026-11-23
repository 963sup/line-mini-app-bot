# 背景工作與排程

<a id="cloud-tasks"></a>

## Cloud Tasks

**狀態：已有摘要佇列接線；其他背景工作按需擴充。**

**用途與最小流程**

將摘要、通知或文件處理交給指定 HTTP worker，控制重試與速率。

Go 建立具名工作 → 以服務身分呼叫 worker → 核對操作狀態 → 提交結果後回覆成功。

**資料、授權與失敗邊界**

工作只帶操作 ID 與必要作用域；worker 重新讀取授權及狀態，以操作鍵防止重複業務效果。建立工作成功不代表業務已完成；Firestore 提交與入列間的失敗窗口須另外設計。

**導入條件**

設定佇列、目標 URL、呼叫身分、期限及重試／失敗處理；區分可重試失敗與永久拒絕。

**驗收**

相同工作重送、錯誤身分、worker 逾時、入列失敗、已提交但回應遺失，均有可確認結果。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/tasks/docs/comp-pub-sub)。[摘要接線](../../../backend/internal/bootstrap/summary.go) · [佇列實作](../../../backend/internal/business/summary/adapter/messaging/cloudtasks/summarization/queue.go)。

<a id="cloud-scheduler"></a>

## Cloud Scheduler

**狀態：按需導入。**

**用途與最小流程**

定時產生日摘要、出勤提醒、檢查或清理工作。

排程以明確時區觸發授權端點或工作入口 → 建立該時段操作 → 由背景工作執行。

**資料、授權與失敗邊界**

排程只負責觸發；以排程名稱及業務時段辨識重複執行，業務日期由服務端計算。

**導入條件**

逐項指定時區、執行時間、呼叫身分、重試與停用方式；清理工作須另有資料保存政策。

**驗收**

同時段重送、服務中斷、跨日及手動補跑不重複通知或清除資料。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/scheduler/docs/overview)。[Tasks](#cloud-tasks) · [Jobs](#cloud-run-jobs)。

<a id="cloud-run-jobs"></a>

## Cloud Run Jobs

**狀態：按需導入。**

**用途與最小流程**

執行批次匯入、匯出、資料回填與統計重建。

授權者建立有範圍的執行請求 → Job 讀取批次 → 分段提交並記錄進度 → 回報成功／部分失敗。

**資料、授權與失敗邊界**

使用專用執行身分；批次按穩定 ID 及版本更新，保留重跑進度，不能覆寫較新的使用者資料。

**導入條件**

明定分片、併行上限、期限、重試、試跑、停止條件與恢復方式；大量資料不塞進互動 API。

**驗收**

中斷後續跑、重複執行、同時線上寫入、非法範圍及部分失敗可追查。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/run/docs/create-jobs)。[資料演進](../../references/ddd/model-evolution.md) · [Scheduler](#cloud-scheduler)。

[Google 服務索引與查證範圍](./)
