# 架構規範

專案預設採模組化單體與輕量 Hexagonal Architecture，以最少必要複雜度維持清楚邊界。

## 原則

- 業務能力依 bounded context 分組；同一程序不代表可以跨模組存取內部資料。
- 依賴指向業務核心。Domain 不依賴 HTTP、LINE、Firestore 或雲端 SDK。
- Application 協調用例，定義它需要的最小 port。
- Adapter 實作 port 並轉譯外部格式、錯誤與生命週期。
- Bootstrap 是 composition root，只建立依賴、註冊入口與管理啟停，不承載業務判斷。
- 同一操作由 HTTP、LINE 或工作程序觸發時，應共用用例而不是複製規則。

## 選擇抽象

先採直接、可讀的實作。只有在責任、變更原因或替換需求已被證明時才抽介面或共用元件。相似程式若屬於不同 bounded context，可以分開保留；不得為消除外觀重複而混合業務語意。

目前不預設引入微服務、通用 Repository framework、完整 CQRS、Event Sourcing 或事件平台。需要獨立部署、不同擴展特性、可靠事件歷史或明確讀寫模型差異時，才透過 ADR 重新評估。

## 跨邊界操作

跨模組合作必須有具名契約與擁有者。同步呼叫要定義期限、取消與錯誤；非同步流程要定義事件版本、去重、排序、重試與停止條件。需要與寫入一致的授權、版本與狀態檢查，必須放在同一交易協調內。

## 變更準則

架構變更必須說明問題、較簡單方案為何不足、依賴方向、相容性、回退方式及驗證證據。跨越多個模組或難以逆轉的決策記錄於 [ADR](../decisions/README.md)。

## 架構與所有權

- MUST preserve approved bounded contexts, ownership, and dependency directions.
- MUST maintain Single Source of Truth for every business rule, contract, data model, and configuration.
- MUST preserve the approved DDD and Hexagonal Architecture boundaries.
- MUST define cross-context interactions through explicit capability boundaries.
- MUST follow the authoritative architecture specifications.
- MUST document material architectural decisions and their trade-offs.

- MUST NOT violate approved bounded-context boundaries or dependency directions.
- MUST NOT introduce circular dependencies.
- MUST NOT access another bounded context's internal implementation or persistence directly.
- MUST NOT write data owned by another bounded context; cross-context changes MUST use an explicit capability contract owned by the responsible context.
- MUST NOT introduce competing architectures for existing capabilities.
- MUST NOT relocate responsibilities without preserving their ownership and contracts.
- MUST NOT introduce infrastructure, transport, framework, or external SDK dependencies into domain models.
- MUST NOT place authoritative business rules in transport adapters or infrastructure.
- MUST NOT create parallel implementations of authoritative business rules.
- MUST NOT couple adapters through another adapter's internal types, persistence models, or provider-specific behavior.
- MUST NOT implement separate authoritative business rules for HTTP, LINE chat, LINE group, MINI App, or other delivery channels.

- SHOULD organize business capabilities around bounded contexts and ownership.
- SHOULD use ubiquitous language consistently across implementation and documentation.
- SHOULD model business behavior through explicit commands, queries, policies, and events where appropriate.
- SHOULD introduce aggregates, entities, and value objects when domain semantics require them.
- SHOULD keep bounded contexts internally cohesive and externally loosely coupled.
- SHOULD use the smallest interface that satisfies the caller's requirements.

- SHOULD NOT split cohesive business capabilities into unnecessarily small modules.
- SHOULD NOT merge unrelated capabilities solely to reduce package counts.
- SHOULD NOT introduce shared business packages without clear ownership.
- SHOULD NOT introduce intermediate services that merely forward calls without adding boundary responsibilities.
- SHOULD NOT create cross-context abstractions that obscure business semantics.
- SHOULD NOT introduce distributed services when an existing modular deployment satisfies requirements.
- SHOULD NOT introduce hidden dependencies through global mutable state or service locators.
- SHOULD NOT move domain-specific behavior into shared packages merely to avoid bounded-context ownership.
- SHOULD NOT expose internal domain implementations as cross-context contracts when an explicit capability contract is sufficient.

- MAY introduce additional domain layers when required by demonstrated business semantics.
- MAY extract shared capabilities when stable ownership and genuine reuse are established.
- MAY introduce new bounded contexts when independently owned business capabilities emerge.
- MAY introduce new architectural patterns when existing patterns cannot adequately satisfy validated requirements.
- MAY use feature flags to support reversible rollout or migration.

## 業務與資源建模

- MUST identify the Actor, Resource, Action, Authorization, State, and Event where applicable.
- MUST model concepts with different authority, lifecycle, or semantics separately.
- MUST assign each business rule to its owning bounded context.
- MUST enforce business invariants before changing resource state.
- MUST define valid state transitions, preconditions, outcomes, and side effects.
- MUST preserve resource identity and lifecycle consistency.

- MUST NOT bypass business invariants or legal state transitions.
- MUST NOT perform state-changing operations through read-only queries.
- MUST NOT allow non-owning modules to independently redefine resource behavior.
- MUST NOT reuse resource identifiers with incompatible semantics.
- MUST NOT publish events representing uncommitted or nonexistent business facts.
- MUST NOT equate Actor with Role, Identity with Membership, Resource with Capability, Action with Event, State with Permission, or Projection with Aggregate.
- MUST NOT use an external platform identifier as canonical domain identity unless the owning domain explicitly defines and controls that identity's meaning and lifecycle.

- SHOULD NOT create orphaned or cyclic relationships in hierarchical resource models.
- SHOULD NOT mutate published template revisions in place.
- SHOULD NOT retroactively modify instantiated objects solely because their source template changed.
- SHOULD NOT infer missing quantities, units, or contractual requirements as verified source facts.
- SHOULD NOT remove source provenance from derived business objects without an explicit retention decision.
