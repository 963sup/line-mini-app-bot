# 儲存模型與資料歸屬

本頁整理目前 Firestore 保存入口及新增資料設計的核對方式。具體欄位與業務規則由[領域規格](../../domains/README.md)維護；通用 schemaVersion、Inbox／Outbox 等仍依[資料契約提案](../../proposals/firestore-data-contracts.md)評估，不能當作全庫現行保證。

## 現行配置與證據

| 資料 | 保存入口與關係 | 原始碼來源 |
|---|---|---|
| Repository | `channels/{channelId}/repositories/{id}`；擁有者綁定另存 `repositoryOwnerBindings`，使用者候選另存 `repositoryUsers` | [Repository store](../../../backend/internal/business/repository/adapter/firestore/store.go) |
| Object | Repository 下的 `objects`；父子關係由 `parentId` 表示；建立收據另存 `objectCreationReceipts` | [Object store](../../../backend/internal/business/object/adapter/firestore/store.go) |
| Wiki | Repository 下的 `wikiPages` 與 `wikiTitles`；每頁有 `revisions` 子集合 | [Wiki store](../../../backend/internal/business/repository/adapter/firestore/wiki.go) |
| 出勤 | 依現行出勤 adapter 的路徑與資料轉譯保存，不因 Repository 文件歸組而搬移 | [出勤 adapter](../../../backend/internal/business/attendance/adapter/firestore/README.md) |
| 瀏覽器 Session | 保存方式與有效期檢查由身分能力維護 | [Session 設計](../identity/browser-session.md) |

這是代表資料的定位表，不是完整集合清單。Firestore 子集合階層不等於 DDD 聚合；父文件存在、授權、唯一性與引用完整性都須查核相應用例及交易。

## 設計一份持久資料

新增或修改保存格式時，列出權威寫入者、身分、作用域、欄位型別、缺值與零值語意、版本、保存期限及可重建來源。區分原始事實、目前快照、查詢投影、收據與稽核，不讓同一事實有兩個不明確的權威來源。

公開 DTO、Domain Entity 與保存格式各有責任。Adapter 負責編碼、解碼及舊格式處理；未知狀態不能猜成成功。數量、時區與精度以領域契約為準，不能讓資料庫型別默默改變語意。

## ID、引用與唯一性

現有共用 ID 編碼由[Firestore key](../../../backend/internal/platform/firestore/key/key.go)維護；變更雜湊輸入或編碼可能改變文件、收據及游標的身分，須先處理歷史相容性。API 與集合命名依[命名規範](../../standards/naming.md)。

引用保存 ID 及必要的作用域，讀取或寫入時由擁有者核對可見性。存在父子路徑不代表有外鍵或級聯刪除。唯一綁定需具體說明保護文件與交易，例如 Repository 的 owner binding；不能用先查詢再無條件新增來宣稱唯一。

## 變更與驗收

列出新舊讀寫版本、歷史資料樣本、回填是否必要、衝突時保留哪一方及可重跑條件。依[模型演進](../ddd/model-evolution.md)先擴充讀端，再按授權處理遷移與收縮。資料夾搬移不授權正式資料搬遷。

驗證編碼／解碼、缺值、非法值、舊格式、跨作用域引用、唯一性競爭與拒絕後狀態。模型測試不能代替儲存整合；正式資料分布仍須另行觀測。

[交易與一致性](consistency.md) · [查詢與索引](querying.md) · [資料生命週期](lifecycle.md) · [儲存索引](README.md)
