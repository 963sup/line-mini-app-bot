# DDD 設計

供 Codex 按問題查閱 DDD 與六邊形架構。先用戰略設計理解業務、通用語言與限界上下文，再用戰術設計表達模型與不變條件，最後用 Port 與 Adapter 保護模型邊界。各頁範例用於解釋概念。

## 如何使用

新設計依[七階段流程](design_process.md)從用例的成功、拒絕與併發反例出發。先辨認值與身分，再找出規則擁有者及需要同時提交的最小狀態，最後推導 Port、Adapter 與驗證。修改既有功能時，從下表選讀對應主題。每個公開入口應能追溯到用例或規則，缺乏證據的邊界標為待決定。

## 按問題查閱

| 問題                                | 閱讀入口                                                                                                   |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| 這項能力是否應獨立成一個上下文？    | [子域](subdomain.md)、[限界上下文](bounded_context.md)                                                     |
| 需要追蹤身分，還是只需要一個值？    | [Entity](entity.md)、[Value Object](value_object.md)                                                       |
| 哪些狀態必須一起成立？              | [Invariant](invariant.md)、[Aggregate](aggregate.md)、[一致性](consistency.md)                             |
| 誰有權操作，狀態允許嗎？            | [領域授權](domain_authorization.md)、[狀態機](state_machine.md)                                            |
| 公開契約如何定義錯誤與版本相容？    | [上下文公開契約](domain_contract.md)、[模型演進](model_evolution.md)                                       |
| 如何探索未知規則、審查模型品質？    | [EventStorming](eventstorming.md)、[模型審查](domain_model_review.md)                                      |
| 建立與持久資料重建放哪裡？          | [Factory](factory.md)、[Repository](repository.md)                                                         |
| 業務判斷放模型還是服務？            | [Entity](entity.md)、[Domain Service](domain_service.md)、[Application Service](application_service.md)    |
| Port、Adapter、交易與錯誤如何劃分？ | [六邊形架構](hexagonal_architecture.md)、[一致性](consistency.md)                                          |
| 資料提交後如何避免遺失發布意圖？    | [Outbox（按需）](outbox_pattern.md#提交與失敗窗口圖)                                                       |
| 跨領域交換什麼，是否需要事件？      | [上下文映射](context_mapping.md)、[整合模式](context_mapping_patterns.md)、[Domain Event](domain_event.md) |

## 設計問題與產出

一項設計應能回答下列問題並交出可查核的證據。新能力逐項評估，局部修正只追受影響項目。資料不足時，記錄需要哪位規則擁有者決定、哪種測試或觀測能分辨選項。

| 決策                                | 入口                                                                          | 可交付的證據                                                             |
| ----------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| 是誰的問題，詞義在哪裡改變？        | [領域探索](domain_modeling.md)、[上下文整合模式](context_mapping_patterns.md) | 一個成功案例、一個有爭議的反例、事實提供方與規則擁有者                   |
| 哪些事必須一起成立？                | [不變條件](invariant.md)、[聚合](aggregate.md)、[一致性](consistency.md)      | 提交邊界、競爭寫入時允許的結果與拒絕後狀態                               |
| 誰呼叫誰，誰可以依賴誰？            | [六邊形架構](hexagonal_architecture.md)、[組裝根](composition_root.md)        | 走通一次入站操作與一次出站呼叫，指出介面擁有者、錯誤映射及具體實作注入處 |
| 重送、部分成功或外部中斷怎麼辦？    | [冪等](idempotency.md)、[交易](transaction.md)、[Outbox](outbox_pattern.md)   | 列出故障點、可觀察狀態、重試責任，以及何時值得採用進階模式               |
| 如何證明替換 Adapter 後仍符合承諾？ | [架構與契約測試](architecture_testing.md)                                     | 相同輸入在不同 Adapter 上的成功、衝突與失敗結果                          |
| 如何讓既有消費者安全過渡？          | [契約與模型演進](model_evolution.md)                                          | 新舊版本相容組合、切換前置條件與安全退回的界線                           |

## 戰略設計

| 概念                                          | 英文術語                 | 設計用途                       |
| --------------------------------------------- | ------------------------ | ------------------------------ |
| [領域](domain.md)                             | Domain                   | 界定業務問題、參與者與成果     |
| [子域](subdomain.md)                          | Subdomain                | 按業務能力劃分問題空間         |
| [核心子域](core_subdomain.md)                 | Core Subdomain           | 找出值得集中投入的差異化能力   |
| [支援子域](supporting_subdomain.md)           | Supporting Subdomain     | 識別需貼合本業流程的支援能力   |
| [通用子域](generic_subdomain.md)              | Generic Subdomain        | 評估可採用成熟方案的能力       |
| [限界上下文](bounded_context.md)              | Bounded Context          | 指定模型、語言與規則的有效範圍 |
| [通用語言](ubiquitous_language.md)            | Ubiquitous Language      | 統一上下文內的詞義與例句       |
| [上下文映射](context_mapping.md)              | Context Mapping          | 調查契約、依賴及整合方式       |
| [上下文映射圖](context_map.md)                | Context Map              | 呈現模型邊界及事實影響方向     |
| [上游／下游](upstream_downstream.md)          | Upstream / Downstream    | 說明契約變更影響哪一方         |
| [上下文整合模式](context_mapping_patterns.md) | Context Mapping Patterns | 選擇契約與模型的合作方式       |

## 戰術設計概念

| 概念                                | 英文術語                   | 設計用途                            |
| ----------------------------------- | -------------------------- | ----------------------------------- |
| [實體](entity.md)                   | Entity                     | 追蹤穩定身分與合法狀態轉換          |
| [值物件](value_object.md)           | Value Object               | 封裝按內容判等的有效值              |
| [不變條件](invariant.md)            | Invariant                  | 明定規則的擁有者與成立時間          |
| [聚合與聚合根](aggregate.md)        | Aggregate / Aggregate Root | 劃分規則與一致性邊界                |
| [工廠](factory.md)                  | Factory                    | 封裝完整物件的建立與初始規則        |
| [儲存庫模式](repository.md)         | Repository                 | 以 Aggregate 為單位載入、重建及保存 |
| [領域授權](domain_authorization.md) | Domain Authorization       | 分開資格、領域規則與權威狀態判定    |
| [狀態機](state_machine.md)          | State Machine              | 定義轉換、Guard、拒絕與終態         |
| [領域服務](domain_service.md)       | Domain Service             | 表達不屬於單一物件的領域決策        |
| [領域事件](domain_event.md)         | Domain Event               | 表達已發生的業務事實與後續處理需求  |

## 用例與架構落地

| 主題                                    | 設計用途                                           |
| --------------------------------------- | -------------------------------------------------- |
| [應用服務](application_service.md)      | 協調用例、授權、模型、保存與結果                   |
| [一致性與交易](consistency.md)          | 決定原子範圍、版本、重送與副作用                   |
| [六邊形架構](hexagonal_architecture.md) | 區分核心、入站／出站 Port、Adapter、組裝與依賴方向 |

六邊形架構是承載業務模型的架構方法，不是 DDD 的同義詞。CQRS、Event Sourcing、微服務與通用事件平台不因採用本索引而成為必要條件。

[API、Go 後端與前端示範檔案樹](hexagonal_architecture.md#go-檔案配置範例)提供按需拆分的參考。

## 進階模式：按需採用

以下模式按問題選用。選擇前先比較較簡單的方案，說明必要性、失敗語意與維運成本。

| 主題                                      | 按需採用的條件                                       |
| ----------------------------------------- | ---------------------------------------------------- |
| [CQRS](cqrs.md)                           | 讀寫模型的差異或效能需求足以抵銷分離成本             |
| [Saga](saga.md)                           | 需要協調跨交易操作、補償與失敗恢復                   |
| [Process Manager](process_manager.md)     | 需要跨事件的持久進度、等待、逾時與續跑               |
| [Transactional Outbox](outbox_pattern.md) | 業務提交後不能因程序中斷遺失發布意圖                 |
| [Event Sourcing](event_sourcing.md)       | 確定以事件序列作狀態權威來源，且能承擔演進與重播成本 |

## 概念細目

### 模型與上下文整合

[上下文公開契約](domain_contract.md) · [領域模型](domain_model.md) · [限界上下文整合](bounded_context_integration.md) · [防腐層](anti_corruption_layer.md)

### 領域模型與規則

[識別](identity.md) · [領域基本型別](domain_primitive.md) · [聚合根](aggregate_root.md) · [領域政策](domain_policy.md) · [規格物件](specification.md) · [領域錯誤](domain_error.md) · [領域事件處理器](domain_event_handler.md)

### 用例及讀寫意圖

[用例](use_case.md) · [領域命令](domain_command.md) · [命令](command.md) · [查詢](query.md)

### 六邊形架構細節

[依賴規則](dependency_rule.md) · [依賴反轉](dependency_inversion.md) · [埠](port.md) · [入站埠](driving_port.md) · [出站埠](driven_port.md) · [轉接器](adapter.md) · [入站轉接器](driving_adapter.md) · [出站轉接器](driven_adapter.md) · [組裝根](composition_root.md)

### 交易與跨邊界一致性

[交易](transaction.md) · [工作單元](unit_of_work.md) · [樂觀並行控制](optimistic_concurrency.md) · [冪等](idempotency.md) · [最終一致性](eventual_consistency.md) · [整合事件](integration_event.md)

### 建模與驗證

[領域建模](domain_modeling.md) · [EventStorming](eventstorming.md) · [模型審查](domain_model_review.md) · [領域測試](domain_testing.md) · [架構測試](architecture_testing.md)

## Mermaid 圖表

[設計推導、上下文、聚合與用例執行圖](design_diagrams.md)。

| 想釐清的問題                         | 圖表入口                                                                                                                                    |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| 哪個上下文提供契約，哪裡轉譯語意？   | [上下文映射圖](context_map.md)、[防腐層](anti_corruption_layer.md#圖解)                                                                     |
| 執行呼叫與程式碼依賴各往哪裡？       | [六邊形架構](hexagonal_architecture.md#執行與程式碼依賴)                                                                                    |
| 哪些狀態共同維護？讀寫模型如何更新？ | [聚合](aggregate.md)、[CQRS](cqrs.md#圖解)                                                                                                  |
| 何時提交與發布？如何處理衝突及重送？ | [交易](transaction.md#提交與發布)、[Outbox](outbox_pattern.md)、[樂觀並行控制](optimistic_concurrency.md#圖解)、[冪等](idempotency.md#圖解) |
| 跨交易失敗如何恢復？誰維護流程進度？ | [Saga](saga.md#圖解)、[Process Manager](process_manager.md#圖解)                                                                            |

## 來源

[Microsoft 戰術 DDD](https://learn.microsoft.com/en-us/azure/architecture/microservices/model/tactical-domain-driven-design)介紹聚合、不變條件與身分引用。[EventStorming](https://www.eventstorming.com/)介紹以業務事件探索領域的工作坊方法。
