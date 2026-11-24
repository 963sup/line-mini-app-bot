# 事件模型

事件表示已發生且值得其他流程知道的事實；Command 表示意圖。兩者不得用相同名稱或 payload 混用。

## 事件契約

跨模組或跨程序事件需定義：

- event name、owner 與版本；
- event ID、occurred time、actor／resource reference；
- 最小必要 payload 與敏感資料限制；
- producer 的提交點；
- consumer 的去重、順序、重試與停止條件；
- schema 相容期與未知版本處理。

## 提交與發布

資料提交與外部發布不是天然原子。若業務要求「寫入成功後最終一定發布」，應使用持久化工作或 outbox 類模式；沒有該要求時，不預先建立通用事件平台。Firestore transaction callback 內不得直接發布外部事件。

## LINE Webhook

Webhook 是外部事件入口：先驗簽，再轉譯成應用命令或事實。平台重送必須可安全處理。HTTP 200 只表示本次接收流程完成，不代表訊息已送達或使用者已讀。

## 背景工作

Cloud Tasks、Jobs、Pub/Sub 或 Workflows 依工作特性選擇。每個工作需要 operation ID、attempt、deadline、租約或去重策略，以及永久失敗的可觀察終態。重試不得更換業務操作識別碼。

## 物件數量異動的 Pub/Sub 投遞

設定 `OBJECT_QUANTITY_PUBSUB_TOPIC=projects/<project>/topics/<topic>` 可啟用數量異動投遞。主題須已建立，執行身分須有發布權限。啟用後，變更總數量時會在同一筆 Firestore 交易寫入物件、版本歷史及待送事件。事件包含頻道、儲存庫、物件、版本與前後數量及單位。未設定時沿用既有數量稽核。

伺服器運行期間每分鐘最多投遞 50 筆至設定的 Google Cloud Pub/Sub 主題，發布成功才移除待送紀錄。發布失敗會保留事件供重試。並行執行或回應遺失可能重送，訂閱端應依訊息屬性 `eventId` 防重。實例停止或無 CPU 時，待送資料留待實例恢復運行後處理。

發布端將一筆事件序列化為 UTF-8 JSON，再以標準 Base64 編碼，放入 Pub/Sub REST 發布請求的 `messages[].data`。同一則訊息的 `attributes.type` 固定為 `object.quantity_changed`，`attributes.eventId` 是該次異動的穩定防重鍵。直接處理 REST 資料時先解 Base64；若訂閱端 SDK 已提供訊息本文的 bytes，直接解析 JSON 即可。[發布實作](../../backend/internal/business/object/adapter/messaging/publisher.go)

以下是解碼後的完整本文範例。欄位名稱大小寫須與範例一致；`Version` 是物件修訂版號，型別為整數，異動前後數量則為字串。[事件型別](../../backend/internal/business/object/domain/event/quantity_changed.go)

```json
{
  "ChannelID": "1234567890",
  "RepositoryID": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "ObjectID": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "Version": 2,
  "Change": {
    "BeforeQuantity": "10",
    "BeforeUnit": "組",
    "AfterQuantity": "12.5",
    "AfterUnit": "組"
  }
}
```

數量事件分頁依文件 ID 輪巡，在嘗試發布前保存目前頻道的游標。單筆內容損壞、發布失敗或確認遺失時保留該筆，繼續其他事件；游標走到底後再從頭重試。同一事件重送沿用相同 eventId。可依「quantity event pending」的事件識別碼定位積壓，修復資料或 IAM 後觀察該筆成功移除。

Summary 以原 job 作為持久化的工作意圖。正常請求經 Cloud Tasks；伺服器另輪巡本頻道既有工作，對超過兩分鐘且沒有有效租約的 pending 工作沿用原 Claim、版本檢查與 Push retry key 執行。重啟後沿用 Firestore 游標。過期未完成的工作在 Claim 時標記 expired；未遭處理即被 TTL 移除的工作仍須依外部監控辨識。重複收到相同訊息收回事件時，既有 tombstone 阻止群組版本重複增加，交易後仍會重試清除來源快取。
