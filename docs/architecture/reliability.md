# 事件、資料與可靠發送

> 狀態：待實作契約。適用於 LINE、Firestore、Garr 的整合與驗收。

## Webhook 與 Inbox

1. 依渠道選擇 secret，對未修改 body 驗簽；驗證後解析，處理空陣列與多事件。
2. 用渠道＋webhookEventId 去重並可靠保存；全部需接收事件保存成功才 ACK。部分失敗後重送仍須安全去重。
3. 依 user/group 分流；未知來源安全忽略並觀測。不要依 isRedelivery 判斷唯一性。
4. 消費者領租約，交易內檢查狀態、更新業務、寫入 Outbox 並標記完成；不得在 transaction callback 呼叫 LINE。
5. 重複與亂序獨立處理；重要狀態根據版本／事件時間決策，Unsend 依保存政策移除或停止使用內容。

## Reply 與 Push

Reply 有獨立優先執行路徑，不等待一般排程。Inbox 去重不保證外部 Reply exactly-once；保存發送狀態、嘗試與結果不明狀態。
Reply token 短效、單次使用，不無限重試；逾期或結果不明不自動改 Push，需業務規則決定。
Push 工作保存 typed target、內容快照／不可變模板版本、固定 retry key、attempt、nextRunAt、leaseOwner、leaseUntil、leaseVersion。
工作完成與續租需驗證租約版本；過期 Worker 不可覆寫新擁有者。重試保留相同 retry key；超出供應商去重有效範圍需人工策略。
API 接受不代表已送達／已讀。沒有端到端 exactly-once 保證。

## Firestore

- server SDK 透過 ADC/IAM，繞過 Security Rules；Go Use Case 做資源授權。
- 依 channel 隔離 users、groups/roles、webhookEvents、outbox、templates/versions、richMenus、auditLogs。
- 個人與群組資料分開；查詢明確帶渠道／資源範圍，不能只靠客戶端傳來的 ID。
- 為 status＋nextRunAt 等工作查詢建立索引；payload 視需求排除索引；避免單文件熱點。
- TTL 只負責最終清理，不能做精準排程；去重保存期需覆蓋產品重試政策。
- 圖片不存 Firestore 文件：需另外選 HTTPS 靜態檔／物件儲存方案；此項尚未選定。

## Worker 與 Garr

Go Worker 持續運行，定期掃描到期工作；限制領取數、程序內 queue 與併發，處理背壓。
Garr 僅包裝 pool、退避與 breaker；自行實作 context 取消、錯誤分類、重試上限與終止狀態。
400／權限等永久錯誤停止重試；429 區分限流與額度，不盲目重試；暫時錯誤有界退避。
多副本共用的渠道／endpoint 限流需另行協調；Garr queue 不取代 Firestore。
SIGTERM 停止領取、限制關閉等待，未完成工作到期可恢復。斷路器開啟時延後工作，不占住 Worker 忙等。

## 安全與觀測

MINI App 使用 LIFF token 後端驗證，不信任 profile；session 使用安全 cookie 與 CSRF 防護。Provider／channel 身分關聯須明確核對。
群組管理角色需驗證綁定，不以首次操作自動授予。日誌不記 secret、token、完整私人訊息。
觀測 Inbox/Outbox 積壓、最老工作時間、Reply 延遲、重試、失敗、breaker 狀態；記錄渠道與追蹤 ID。
部署唯一使用 Fly.io（ADR 0001）；需確認 Worker 常駐 Machine、Google ADC/IAM 與區域。現有 HTTP 配置不代表 Worker 已可運行。

依據：[Firestore 交易](https://docs.cloud.google.com/firestore/native/docs/manage-data/transactions)、[IAM](https://docs.cloud.google.com/firestore/native/docs/security/iam)、[Garr](https://github.com/line/garr)、[LINE 能力矩陣](../integrations/line.md)。
