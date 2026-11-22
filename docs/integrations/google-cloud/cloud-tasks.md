# Cloud Tasks 簡要規格

**狀態：已有摘要佇列接線；其他背景工作按需擴充。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

將摘要、通知或文件處理交給指定 HTTP worker，控制重試與速率。

Go 建立具名工作 → 以服務身分呼叫 worker → 核對操作狀態 → 提交結果後回覆成功。

## 資料、授權與失敗邊界

工作只帶操作 ID 與必要作用域；worker 重新讀取授權及狀態，以操作鍵防止重複業務效果。建立工作成功不代表業務已完成；Firestore 提交與入列間的失敗窗口須另外設計。

## 導入條件

設定佇列、目標 URL、呼叫身分、期限及重試／失敗處理；區分可重試失敗與永久拒絕。

## 驗收

相同工作重送、錯誤身分、worker 逾時、入列失敗、已提交但回應遺失，均有可確認結果。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/tasks/docs/comp-pub-sub)，查證範圍見[整合總覽](README.md#sources)。[摘要接線](../../../backend/internal/bootstrap/summary.go) · [佇列實作](../../../backend/internal/business/summary/adapter/messaging/cloudtasks/summarization/queue.go)。

[Google 服務索引](README.md)
