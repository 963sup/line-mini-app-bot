# 修改領域行為

授權、狀態轉換、重送、併發或資料寫入規則有變時使用。先取得需求與預期的成功、拒絕和失敗狀態。

1. 從[領域規格](../../domains/README.md)選功能，依[業務操作模型](../../architecture/contracts/business-operations.md)定位 Actor、Resource、Action 與 Authority。
2. 從使用者入口追到 use case、聚合或 policy，再核對 port、adapter 及交易邊界。邊界存在疑問時，從 [DDD 問題索引](../../architecture/ddd/README.md#按問題查閱)按題選讀。
3. 依[業務符合性流程](../workflow/implementation.md#business-operation-compliance)同步規格、判斷與實作。影響 API 時，也走[契約配方](change-api.md)。
4. 依[驗收矩陣](../verification/commands.md#變更驗收矩陣)執行既有適用案例，記錄能支持哪些授權、重送、交易或併發結論。

完成時指出改動的權威規格、Go 符號、寫入位置與已驗證的結果。若缺少整合證據，寫出仍須核對的交易或併發行為。[出勤查詢與補正規格](../../domains/attendance/queries.md)

[操作配方](README.md)
