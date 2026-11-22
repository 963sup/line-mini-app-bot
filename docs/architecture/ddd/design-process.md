# 從領域探索到六角形架構

以「員工補正一筆班次」為例，前五步釐清問題、語言與合作邊界，後兩步把業務決策落到程式。測試或業務訪談若揭露不同語意，可以回頭修正前面的判斷。

| 階段 | 在補正案例要回答的問題 | 可核對的產出 |
|---|---|---|
| 1. Domain Discovery（領域探索） | 誰提出補正？何時生效？ | 以 [EventStorming](eventstorming.md) 探索申請、取消、審核與結果，核對成功與拒絕案例。 |
| 2. Subdomain Classification（子域分類） | 哪些能力構成產品優勢？哪些是支援或可採購的能力？ | 列出分類理由及尚待產品驗證的假設。 |
| 3. Bounded Context Design（限界上下文設計） | 誰提供可信身分？誰決定此人能否核准這筆班次？ | 畫出身分與出勤的責任邊界，列明各自擁有的規則。 |
| 4. Ubiquitous Language（通用語言） | 「補正申請」「審核決定」「生效紀錄」分別指什麼？ | 用相同詞彙描述介面、領域行為與驗收案例。 |
| 5. Context Mapping（上下文映射） | 出勤接收哪些身分事實？變更會影響哪一方？ | 畫出有方向及契約標籤的 Context Map，記錄整合方式。 |
| 6. Tactical Design（戰術設計） | 哪個 Entity 有穩定身分？哪些值按內容相等？哪些規則需要同時成立？複雜物件如何完整建立？ | 區分 [Entity](entity.md) 與 [Value Object](value-object.md)，依 [Invariant](invariant.md) 評估 [Aggregate](aggregate.md) 邊界；選擇 [Factory](factory.md)、[Repository](repository.md)、[Domain Service](domain-service.md) 與必要的 [Domain Event](domain-event.md)。 |
| 7. Hexagonal Architecture（六角形架構） | 哪一層執行用例？如何在最新資料上做決策並原子寫回？ | 以 [Application Service](application-service.md) 協調用例，依[六邊形架構](hexagonal-architecture.md)定義 Port／Adapter 與依賴方向，並明列[一致性](consistency.md)、併發及重送的驗收需求。 |

完成各階段後依[模型審查](domain-model-review.md)核對規則擁有者、反例、公開契約與設計驗收證據。

核對專案已實作的行為與下一批設計目標，見[出勤規格](../../domains/attendance/README.md)。[DDD 設計索引](README.md)
