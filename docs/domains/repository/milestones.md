# 里程碑

<a id="里程碑"></a>

[進度顯示「已關閉 X/Y 個問題」，依所屬 Issue 的關閉數與總數計算。里程碑本身由建立者手動關閉或重開。正式環境驗收須有實際證據。](labels.md#標籤引用切片)

<a id="操作契約"></a>

## 操作契約

| 項目 | 規則 |
| --- | --- |
| Actor／Resource／Action | Session／Bearer 核驗本人；同一啟用儲存庫中的 Milestone；列出、建立、詳情與更新 |
| Context／Authority | 核驗當下群組成員資格；建立者才能編輯 |
| Policy／Decision | 標題 1–120 字元、描述最多 8000 字元；可選 YYYY-MM-DD 曆日 |
| Result | 回傳內容、版本與更新時間；無事項時顯示「尚無事項」 |
| 重送／併發 | 建立鍵同內容回同一 ID，異內容回 409；更新檢查 expectedVersion，衝突回 409；無變化不增加版本 |
| Event／Audit | 保存建立者與時間、更新操作者／版本／時間；無外部通知 |

<a id="資料與相容邊界"></a>

## 資料與相容邊界

- 核心位於 backend/internal/business/repository/milestones，Firestore 使用既有 Repository Store；資料在 channel／repository 下的 milestones，建立收據使用對應功能的 receipt collection。不占共用資源編號，不更動既有事項／討論收據。

- 舊事項缺少 milestoneId 時視為未歸屬。事項建立收據雜湊保持相容。寫入交易檢查容器啟用、群組綁定、作者、版本與引用範圍，全部讀取先於寫入。

- 進度由當前來源事項計算，不複製事項狀態；查詢成本隨資料量增長，不自行加入計數器或排程。

<a id="契約與驗收入口"></a>

## 契約與驗收入口

- OpenAPI 來源 · 事項規格 · 驗收矩陣。文件規則與歷史通過紀錄均不能代替本次變更或目標環境的驗收。

  - [OpenAPI 來源](../../../api/contract/repository/milestones.yaml)

  - [事項規格](issue/design.md)

  - [驗收矩陣](../../development/tooling/testing/commands.md)
