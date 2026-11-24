# 實體（Entity）

Entity 以身分延續來識別。同一身分的狀態可以改變；欄位相同但身分不同，仍是兩個實體。是否需要 Entity，取決於業務是否追蹤「同一個對象」，不取決於資料表是否有主鍵。

## 適用判斷與責任

先定義識別範圍、生命週期與合法狀態轉換，再決定識別碼。跨渠道的相同字串不一定是同一身分；Aggregate 內部實體也可能只需要局部識別。

Entity 應透過具名業務行為改變狀態，維持自己負責的規則。只有 Aggregate Root 是整個一致性邊界的入口，不能因為 Entity 有 ID 就讓外部繞過根修改它。Repository 保存的是模型需要的狀態，不替實體定義業務身分。

## 設計例子

例如，同一關係的角色可由 member 改成 owner，關係的識別不必改變；移除後重新加入是否沿用身分，必須由領域明確定義。

## 設計產出與核對

記錄 ID 的範圍、產生者、生命週期、合法行為及版本語意。核對同 ID 的不同狀態仍屬同一實體，並明確處理重建舊狀態與重複建立。

概念參考：[Microsoft Learn：戰術 DDD](https://learn.microsoft.com/en-us/azure/architecture/microservices/model/tactical-domain-driven-design)。僅採概念，不採其微服務部署前提。

[Value Object](value_object.md) · [Aggregate](aggregate.md) · [DDD 設計索引](design_guide.md)
