# 整合事件（Integration Event）

面向其他上下文的已發生事實與公開契約。

## 設計時要確認

為事件 ID、版本、最小資料和去重訂約；從已提交狀態產生並容許重送。

## 示例

InvitationAccepted 可對外提供處理必要的識別，內部拒絕原因無須成為公開 schema。

## 延伸閱讀

[domain-event](domain_event.md) · [outbox-pattern](outbox_pattern.md) · [bounded-context-integration](bounded_context_integration.md) · [DDD 設計索引](design_guide.md)
