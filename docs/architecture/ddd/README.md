# DDD 戰略設計

這組文件提供跨領域的設計方法。各領域採用的詞彙、現行規則與待確認事項，記錄在[領域規格](../../domains/README.md)。核心、支援與通用子域文件中的例子是分類假設。

| 類別 | 英文術語 | 核心用途 |
|---|---|---|
| [領域](domain.md) | Domain | 界定團隊要理解的業務問題、參與者與成果。 |
| [子域](subdomain.md) | Subdomain | 按業務能力劃分問題空間。 |
| [核心子域](core-subdomain.md) | Core Subdomain | 找出能建立產品競爭優勢、值得集中投入的能力。 |
| [支援子域](supporting-subdomain.md) | Supporting Subdomain | 識別支撐核心成果、需要貼合本業流程的能力。 |
| [通用子域](generic-subdomain.md) | Generic Subdomain | 辨認可用成熟通用方案解決的業務能力。 |
| [限界上下文](bounded-context.md) | Bounded Context | 界定一套模型、語言與規則在哪裡有效。 |
| [通用語言](ubiquitous-language.md) | Ubiquitous Language | 讓業務、產品、工程與測試在同一上下文使用一致的詞義。 |
| [上下文映射](context-mapping.md) | Context Mapping | 調查上下文如何依賴彼此，並協商整合方式。 |
| [上下文映射圖](context-map.md) | Context Map | 用圖呈現模型邊界、依賴及關係。 |
| [上游／下游](upstream-downstream.md) | Upstream / Downstream | 描述某項模型或契約的變更會影響哪一方。 |
| [上下文整合模式](context-mapping-patterns.md) | Context Mapping Patterns | 選擇符合模型差異、協作能力與維護成本的整合方式。 |

[從領域探索到六角形架構的七階段流程](design-process.md)
