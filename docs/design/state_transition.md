# 狀態轉換模型

有生命週期的資源必須以明確狀態與合法轉換描述，不用零散布林值或 UI 文案推定。

## 轉換契約

每個轉換需要：

| 欄位          | 內容                       |
| ------------- | -------------------------- |
| Command       | actor 的具名意圖           |
| Current state | 判斷時讀到的權威狀態與版本 |
| Guard         | 授權、資料完整性與業務條件 |
| Next state    | 成功後唯一合法狀態         |
| Result        | 呼叫端可觀察的資料與錯誤   |
| Side effects  | 提交後才執行的外部動作     |
| Evidence      | 必要事件、收據或稽核       |

## 並行與重送

狀態轉換使用版本或交易保護，避免舊畫面覆寫新結果。相同 idempotency key 與相同內容可以回覆原結果；相同鍵搭配不同內容必須拒絕。回應遺失時以原鍵重試或查詢結果，不把 timeout 當成「未執行」。

## 失效與刪除

停用、封存、撤銷、過期與實體刪除是不同轉換。業務失效通常先阻止新操作並保留必要歷史；資料清除需另定保存、子集合、外部引用與恢復規則。

## 測試

每個狀態模型至少驗證允許轉換、禁止轉換、重複命令、舊版本、交易失敗及副作用失敗。測試同時確認拒絕後沒有部分寫入或錯誤狀態。

## 出勤補正審核

補正清單及審核詳情回傳目前查詢身分可執行的操作。本人清單的待審與過期申請可撤回。審核者可核准待審申請，也可駁回待審或過期申請。畫面依 `canApprove`、`canReject`、`canCancel` 顯示操作；送出後仍在交易內核對資格與最新資料。

[補正契約](../../api/contract/attendance/corrections.yaml) · [審核契約](../../api/contract/attendance/approvals.yaml)

## 專案引用結果核對

`GET /api/v1/owners/{ownerId}/projects/{id}/reference-state` 供已登入的專案建立者核對目前引用狀態。後端核驗 Project owner 範圍與建立者，再於交易內重查 owner、Project 及目標引用，不讀取來源內容，因此來源權限失效時仍可核對。

輸入為 `desiredPresent` 與互斥的兩種識別：`itemId`，或完整來源組合 `sourceRepositoryId`、`sourceType`、`sourceId`（子物件另帶 `parentObjectId`）。混用、缺漏或格式無效回傳 `invalid_input`；非建立者拒絕存取。回傳 `present` 與 `matches`，只代表讀取當下狀態，不是原寫入的成功收據，也不保證後續未被其他操作改變。

這是唯讀操作，不新增業務事件、不寫入資料；回應禁止快取。前端在符合預期後解除待確認狀態並重新讀取專案，不再逐頁掃描引用推斷原操作結果。未符合或查詢失敗時保留原請求供重試。

契約見 [references.yaml](../../api/contract/project/references.yaml)，核心見 [references.go](../../backend/internal/business/project/domain/entity/references.go)。既有引用授權、HTTP 與前端流程檢查涵蓋此路徑；真實 Firestore 交易與 LINE 裝置驗收仍需另行執行。

## 摘要重送與訊息收回

- 模型呼叫完成但結果尚未保存時若程序中斷，可能重新呼叫模型；不宣稱模型費用 exactly-once。不同摘要指令是不同工作，第一版沒有跨請求快取或頻率限制。

- 收到收回事件時覆寫為不含原文的 tombstone（保留48小時防止延遲原訊息恢復），使舊版本工作失效並清除相關結果；舊模型結果不能寫回。 現行入口不接受「關閉摘要／開啟摘要」，也不會因這兩個指令切換版本或建立清理工作。worker 保留處理既有 purge 工作的相容分支；該分支清理較舊 Epoch 的訊息與工作子集合，不假設刪除父文件可連帶刪除。 發送前重查版本；已開始的外部 Push、已送出的 LINE 訊息或已提交給模型的請求無法藉由後續收回撤銷。

[摘要工作](../../backend/internal/business/summary/application/summarization/summary.go) · [資料保存](../../backend/internal/business/summary/adapter/firestore/summary.go)

## 里程碑更新與進度計算

| 項目                    | 規則                                                                                        |
| ----------------------- | ------------------------------------------------------------------------------------------- |
| Actor／Resource／Action | Session／Bearer 核驗本人；同一啟用儲存庫中的 Milestone；列出、建立、詳情與更新              |
| Context／Authority      | 核驗當下群組成員資格；建立者才能編輯                                                        |
| Policy／Decision        | 標題 1–120 字元、描述最多 8000 字元；可選 YYYY-MM-DD 曆日                                   |
| Result                  | 回傳內容、版本與更新時間；無事項時顯示「尚無事項」                                          |
| 重送／併發              | 建立鍵同內容回同一 ID，異內容回 409；更新檢查 expectedVersion，衝突回 409；無變化不增加版本 |
| Event／Audit            | 保存建立者與時間、更新操作者／版本／時間；無外部通知                                        |

- 進度由當前來源事項計算，不複製事項狀態；查詢成本隨資料量增長，不自行加入計數器或排程。

[里程碑契約](../../api/contract/repository/milestones.yaml)
