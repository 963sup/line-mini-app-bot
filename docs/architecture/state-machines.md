# 事件與發送狀態機

> 狀態：技術契約；外部 API 不提供與 Firestore 跨系統原子交易，不保證 exactly-once。

## Inbox

```mermaid
stateDiagram-v2
  [*] --> pending: durable create-if-absent
  pending --> processing: 原子領取租約
  processing --> completed: 業務變更與 Outbox 同交易提交
  processing --> ignored: 已識別為範圍外或無操作
  processing --> retry_wait: 暫時失敗
  retry_wait --> processing: 到期重新領取
  processing --> pending: 租約過期並以版本檢查回收
  processing --> failed: 永久失敗或上限
```

所有交易先讀後寫；callback 無 LINE 呼叫。重複 event ID 回接收成功，不重設既有狀態。
Inbox completed 只表示處理與通知意圖已提交，不表示通知已送達；Reply 記錄另有狀態。

## Outbox

```mermaid
stateDiagram-v2
  [*] --> pending
  pending --> leased: 原子 claim
  leased --> sending: 保存發送意圖與固定 retry key
  leased --> pending: 租約過期，尚未進入 sending
  sending --> accepted: API 接受或確認相同 retry key 已接受
  sending --> retry_wait: 明確可重試失敗
  sending --> unknown: 超時或 crash，可能已接受
  sending --> failed: 永久錯誤
  retry_wait --> leased: nextRunAt 到期
  unknown --> retry_wait: 仍在可安全去重範圍，沿用 key
  unknown --> manual_review: 超出去重範圍或無法判斷
  retry_wait --> failed: 超過上限
```

sending 過期不得當成未發送的 pending。先持久化 sending 再呼叫 API；租約版本只保護資料更新，不能取消已在外部執行的請求。
避免 lease 在呼叫期間失效：外部 timeout 小於剩餘租約，必要時續租；失去租約者不得提交完成。
409 只有在 endpoint 的重複 retry key 語義確認後才能視為 accepted；保存 accepted request ID，不把所有 409 當成功。
Reply 沒有 Push retry key 的同等保障：not_attempted → attempting → accepted/rejected/unknown；unknown 不自動重送或改 Push。
具體 retry 次數、延遲、去重期限由 endpoint 現行規則與 P05 確認，未確認不得開正式自動重送。

## Webhook 時序

```mermaid
sequenceDiagram
  participant L as LINE
  participant H as Go Webhook
  participant F as Firestore
  participant W as Dispatcher
  participant U as Direct / Group Use Case
  L->>H: 原始 body + signature
  H->>H: 驗簽、結構檢查
  H->>F: create Inbox（事件 ID 去重）
  alt 保存失敗
    H-->>L: 503；已保存部分可安全重送
  else 保存成功
    H-->>L: 200（非業務完成）
    W->>F: claim Inbox 租約
    W->>U: 事件與可信來源資料
    U->>F: 交易：業務變更 + Outbox + 完成狀態
    Note over U,F: 不在 callback 內呼叫 LINE
  end
```

Reply 由獨立優先路徑處理，受 token 期限約束；不能用定期慢掃描保證即時 Reply。實作時以延遲測試選定喚醒／掃描策略。
