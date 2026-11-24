# DDD 與六邊形架構參考

本目錄收錄領域驅動設計（DDD）與六邊形架構的概念參考。以下按用途列出全部主題文件；設計時可先從下方的設計索引按問題選讀。專案現行模型以[設計文件](../design/)為準，依賴方向以[架構邊界](../architecture/boundaries.md)為準。

## 開始設計

- [設計索引](design-guide.md) · [七階段設計流程](design-process.md) · [設計圖解](design-diagrams.md) · [模型審查](domain-model-review.md)

## 戰略設計與上下文整合

- 問題空間：[領域](domain.md) · [子域](subdomain.md) · [核心子域](core-subdomain.md) · [支援子域](supporting-subdomain.md) · [通用子域](generic-subdomain.md)
- 探索與語言：[EventStorming](eventstorming.md) · [限界上下文](bounded-context.md) · [通用語言](ubiquitous-language.md)
- 邊界與合作：[上下文映射圖](context-map.md) · [上下文映射](context-mapping.md) · [上下文整合模式](context-mapping-patterns.md) · [上游／下游](upstream-downstream.md) · [限界上下文整合](bounded-context-integration.md) · [防腐層](anti-corruption-layer.md)

## 領域模型與規則

- 模型與值：[領域模型](domain-model.md) · [領域建模](domain-modeling.md) · [實體](entity.md) · [識別](identity.md) · [值物件](value-object.md) · [領域基本型別](domain-primitive.md)
- 一致性邊界：[聚合](aggregate.md) · [聚合根](aggregate-root.md) · [不變條件](invariant.md) · [領域政策](domain-policy.md) · [規格物件](specification.md)
- 行為與保存：[領域服務](domain-service.md) · [領域授權](domain-authorization.md) · [狀態機](state-machine.md) · [工廠](factory.md) · [儲存庫模式](repository.md)
- 結果與事件：[領域錯誤](domain-error.md) · [領域事件](domain-event.md) · [領域事件處理器](domain-event-handler.md) · [整合事件](integration-event.md)

## 用例與公開契約

- [用例](use-case.md) · [應用服務](application-service.md) · [領域命令](domain-command.md) · [命令](command.md) · [查詢](query.md) · [上下文公開契約](domain-contract.md)

## 六邊形架構與驗證

- 結構與依賴：[六邊形架構](hexagonal-architecture.md) · [依賴規則](dependency-rule.md) · [依賴反轉](dependency-inversion.md) · [組裝根](composition-root.md)
- 埠與轉接器：[埠](port.md) · [入站埠](driving-port.md) · [出站埠](driven-port.md) · [轉接器](adapter.md) · [入站轉接器](driving-adapter.md) · [出站轉接器](driven-adapter.md)
- 測試：[領域測試](domain-testing.md) · [架構測試](architecture-testing.md)

## 交易與可靠性

- [一致性](consistency.md) · [交易](transaction.md) · [工作單元](unit-of-work.md) · [樂觀並行控制](optimistic-concurrency.md) · [冪等](idempotency.md) · [最終一致性](eventual-consistency.md) · [交易式待送紀錄](outbox-pattern.md) · [模型演進](model-evolution.md)

## 按需採用的模式

- [CQRS](cqrs.md) · [事件溯源](event-sourcing.md) · [Saga](saga.md) · [流程管理器](process-manager.md)
