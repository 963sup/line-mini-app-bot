# 資料儲存

本目錄維護資料配置、交易、查詢、索引與生命週期。具體業務欄位及授權以[領域規格](../../domains/README.md)為準，模型邊界依 [DDD 設計](../ddd/README.md)，保存細節由所屬 Firestore adapter 負責。

## 按問題查閱

| 問題 | 內容 |
|---|---|
| 資料存在哪裡？誰有權寫入？ | [儲存模型與歸屬](data-model.md)：現行路徑、ID、引用、序列化及演進 |
| 哪些檢查與寫入必須一起提交？ | [交易與一致性](consistency.md)：讀寫集合、版本、重試、冪等及結果未知 |
| 如何查詢、分頁及部署索引？ | [查詢與索引](querying.md)：作用域、游標、正式索引及成本 |
| 過期、刪除、保存與恢復怎麼辦？ | [資料生命週期](lifecycle.md)：TTL、孤兒資料、保存決策及恢復驗收 |
| 如何連線及使用 Firestore client？ | [Firestore 整合](firestore.md)：執行身分、client、設定與唯讀檢查 |

## 現況與設計界線

現行保存實作由各頁連到原始碼；索引及 Session TTL 宣告集中於 [firestore.indexes.json](../../../firestore.indexes.json)。是否已在雲端套用、完成備份或通過正式負載驗收，須另查目標環境。

通用 schemaVersion、Inbox／Outbox 與統一保存政策仍是[資料契約提案](../../proposals/firestore-data-contracts.md)的討論範圍，不能因本目錄補齊而宣稱已實作。本次文件整理不執行資料搬遷、清除或部署。

## 維護與驗收

異動保存行為時，同步核對領域規格、Adapter、公開契約及歷史資料相容性。交易與競爭驗收依[Firestore 整合流程](../../development/verification/commands.md#firestore-integration)，索引與正式設定依[發布驗收](../../guides/release/release-acceptance.md)。文件檢查不代替資料庫或恢復演練。

[適用工作指引](../AGENTS.md) · [架構索引](../README.md)
