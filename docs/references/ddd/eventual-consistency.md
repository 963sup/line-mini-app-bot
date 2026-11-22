# 最終一致性（Eventual Consistency）

在更新停止且傳播可繼續時，副本或投影朝權威狀態收斂。

## 設計時要確認

定義允許延遲、順序、重播與補齊方式；立即需要的規則讀取具相應保證的來源。

## 示例

邀請列表可能短時間仍顯示待接受，但再次執行命令以權威狀態判斷。

## 延伸閱讀

[consistency](consistency.md) · [query](query.md) · [integration-event](integration-event.md) · [DDD 設計索引](README.md)
