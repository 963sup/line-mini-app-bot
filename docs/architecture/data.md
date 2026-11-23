# 資料儲存

資料擁有者、業務欄位與授權由[領域規格](../domains/)維護。聚合、一致性邊界及資料演進依 [DDD 設計](../references/ddd/design-guide.md)判定。Adapter 負責保存格式與相容性。

- [Firestore 整合](../integrations/firestore/)維護文件配置、交易、查詢、索引與資料生命週期。
- [Cloud Storage 整合](../integrations/storage/)維護檔案、物件版本、IAM 與附件協調。

## 現況與設計界線

現行保存實作由各頁連到原始碼；索引及 Session TTL 宣告集中於 [firestore.indexes.json](../../firestore.indexes.json)。是否已在雲端套用、完成備份或通過正式負載驗收，須另查目標環境。

通用 schemaVersion、Inbox／Outbox 與統一保存政策仍是[資料契約提案](../proposals/firestore-data-contracts.md)的討論範圍，不能因本目錄補齊而宣稱已實作。本次文件整理不執行資料搬遷、清除或部署。

## 維護與驗收

異動保存行為時，同步核對領域規格、Adapter、公開契約及歷史資料相容性。交易與競爭驗收依[Firestore 整合流程](../development/tooling/testing/commands.md#firestore-integration)，索引與正式設定依[發布驗收](../operations/release-acceptance.md)。文件檢查不代替資料庫或恢復演練。

[適用工作指引](AGENTS.md) · [架構索引](./)
