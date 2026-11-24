# 領域事件處理器（Domain Event Handler）

回應同一模型中已發生事實的處理器。

## 設計時要確認

明確決定同步處理是否加入原交易；提交後處理須考慮重送與失敗恢復。

## 示例

處理 InvitationAccepted 可排入持久發布意圖；交易 callback 內避免直接寄通知。

## 延伸閱讀

[domain-event](domain-event.md) · [integration-event](integration-event.md) · [outbox-pattern](outbox-pattern.md) · [DDD 設計索引](design-guide.md)
