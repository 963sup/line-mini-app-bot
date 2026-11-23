# 儲存交易與一致性

本頁將業務不變條件對應到持久化保證。模型邊界見[聚合](../../references/ddd/aggregate.md)，操作責任見[業務操作模型](../../contracts/business-operations.md)。現行 Firestore Go SDK 版本由 [go.mod](../../../backend/go.mod)固定，不在文件另立版本設定。

## 從規則推導讀寫集合

每個重要寫入先列出判定所需的文件、版本與外部事實，再列出必須一起提交的結果。讀取時合法不代表提交時仍合法；需要一致保護的資料必須參與交易或其他明確的衝突控制。

| 代表操作              | 需核對的保存保證                               | 現行來源                                                                                     |
| --------------------- | ---------------------------------------------- | -------------------------------------------------------------------------------------------- |
| 建立／更新 Repository | 唯一擁有者綁定與容器一致；舊事件不能覆寫新觀察 | [Repository store](../../../backend/internal/business/repository/adapter/firestore/store.go) |
| 建立 Object           | 容器與父項有效，重送收據與新物件一起提交       | [Object store](../../../backend/internal/business/object/adapter/firestore/store.go)         |
| 更新 Object           | 目前版本及進度規則成立，物件與變更紀錄一起保存 | [Object store](../../../backend/internal/business/object/adapter/firestore/store.go)         |
| 保存 Wiki             | 頁面、標題占用與修訂保持一致，過期版本拒絕     | [Wiki store](../../../backend/internal/business/repository/adapter/firestore/wiki.go)        |

表格是原始碼定位，未代表本次已執行上述整合驗收。

## Firestore 交易約束

依[官方交易文件](https://firebase.google.com/docs/firestore/manage-data/transactions)，交易先讀後寫，callback 可能重跑。callback 內只處理可重算的讀取、判定與待提交寫入；通知、HTTP 呼叫及其他外部副作用須在交易外協調。交易重試不等於業務請求具備冪等性。

每次重跑使用該次讀到的最新狀態；不要把上一次 callback 的成功輸出當成提交結果。建立 ID、操作鍵與判定時間是否固定，須依操作語意決定並維持重試一致。資源由誰建立、關閉及取消見[Firestore 整合](connection.md)。

## 衝突、重送與結果未知

- 預期版本不符時依契約拒絕，不能自動把使用者的舊編輯覆寫到最新版。
- 相同操作鍵與相同內容可依收據回覆原結果；換內容應依功能契約拒絕，不把鍵視為任意覆寫許可。
- 提交後回應遺失時，呼叫端可能不知道結果。以原鍵重試或查詢操作結果，不能把逾時一律解讀成未寫入。
- LINE 資格等外部事實不屬於 Firestore 的原子快照。說明觀察時點與失敗拒絕方式，不宣稱跨服務原子性。

跨聚合本機交易可作有理由的取捨；跨服務一致性另依[最終一致性](../../references/ddd/eventual-consistency.md)設計。Outbox 等可靠發布模式須有明確需求才採用，不推定已在全庫實作。

## 驗收與設計產出

交付讀寫集合、交易內判定、提交點、版本條件、重送鍵、回應遺失後續行為及外部副作用責任。驗證兩人同版本更新、同鍵同內容／不同內容建立、容器停用競爭、callback 重試及部分失敗後無部分提交。依[Firestore 整合驗收](../../development/tooling/testing/commands.md#firestore-integration)執行；emulator 不代表正式環境的 IAM、索引與負載均已驗證。

[資料配置](data-model.md) · [冪等](../../references/ddd/idempotency.md) · [交易](../../references/ddd/transaction.md) · [儲存索引](./)
