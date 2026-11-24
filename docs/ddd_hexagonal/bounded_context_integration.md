# 限界上下文整合（Bounded Context Integration）

透過公開能力交換事實或請求行為。

## 設計時要確認

以[上下文公開契約](domain_contract.md)記錄語意、錯誤、版本與相容組合。先定契約擁有者、時效與失敗語意；同步查詢適合當下需要回答，非同步適合容許延遲的傳播。

## 示例

Team 查 Organization 資格時把不可用、查無資格與結果過期分開。

## 延伸閱讀

[context-map](context_map.md) · [anti-corruption-layer](anti_corruption_layer.md) · [integration-event](integration_event.md) · [DDD 設計索引](design_guide.md)
