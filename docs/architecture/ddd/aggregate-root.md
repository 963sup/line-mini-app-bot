# 聚合根（Aggregate Root）

管理聚合內狀態變更並保護不變條件的入口實體。

## 設計時要確認

外部透過根發起變更，跨聚合保留識別；以衝突策略防止兩次更新各自通過同一條件。

## 示例

接受邀請經 Invitation.Accept，呼叫端無法直接寫入其狀態欄位。

## 延伸閱讀

[aggregate](aggregate.md) · [invariant](invariant.md) · [optimistic-concurrency](optimistic-concurrency.md) · [DDD 設計索引](README.md)
