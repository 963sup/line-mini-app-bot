# 聚合根（Aggregate Root）

管理聚合內狀態變更並保護不變條件的入口實體。

## 設計時要確認

外部透過根發起變更，跨聚合保留識別；以衝突策略防止兩次更新各自通過同一條件。

## 示例

若建模已證明 Invitation 是聚合根，可由 Invitation.Accept 封裝邀請自身的合法狀態轉換，呼叫端無法直接寫入其狀態。若 Invitation 由另一個根管理，外部異動須由該根發起。接受邀請是否還要原子建立成員資格，須由完整用例的不變條件決定。

## 延伸閱讀

[aggregate](aggregate.md) · [invariant](invariant.md) · [optimistic-concurrency](optimistic-concurrency.md) · [DDD 設計索引](design-guide.md)
