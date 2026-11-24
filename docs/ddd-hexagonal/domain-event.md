# 領域事件（Domain Event）

Domain Event 表示領域中已發生且值得讓其他行為知道的事實，以過去式業務語意命名。它不是要求某件事發生的命令，也不因出現 event 型別就保證可靠發送。

## 區分不同紀錄

| 類型              | 用途與成功邊界                                             |
| ----------------- | ---------------------------------------------------------- |
| Command           | 請求執行操作，可以被拒絕                                   |
| Domain Event      | 模型完成合法轉換所形成的事實；對外視為已提交須等待交易成功 |
| Integration Event | 對其他上下文發布已提交事實的版本化契約，內容需另行轉譯     |
| Audit             | 追溯操作者、對象與結果，也可記錄拒絕；不等同業務事件       |
| 技術 Log          | 診斷執行狀況，不作領域權威來源                             |

## 交付與一致性

模型可以回傳待處理事件，由 Application 安排處理；不在 Entity 中直接呼叫訊息 broker。若 handler 必須影響同次提交的成功，須明列交易範圍；若提交後才處理，須定義失敗、重送、順序及重複消費。

記憶體事件在程序終止時可能遺失；可靠性需求確定後才評估持久待送紀錄等機制。Domain Event 不強制使用非同步、Event Sourcing、通用 dispatcher 或 Outbox。

## 設計例子

未來 MemberRemoved 可以觸發通知或讀模型更新，但撤權若承諾立即生效，不能只等待非同步事件清理權限。正確授權仍須讀取目前有效資格。

## 設計產出與核對

列出事件擁有者、發生時間、提交時間、必要識別、契約版本、消費方與投遞承諾；不攜帶不必要個資。用例直接呼叫已足夠時，不為形式增加事件。

概念參考：[Microsoft Learn：領域事件](https://learn.microsoft.com/en-us/dotnet/architecture/microservices/microservice-ddd-cqrs-patterns/domain-events-design-implementation)。不直接採用其 .NET mediator 或微服務前提。

[Application Service](application-service.md) · [一致性](consistency.md) · [Outbox（按需）](outbox-pattern.md) · [Event Sourcing（按需）](event-sourcing.md) · [DDD 設計索引](design-guide.md)
