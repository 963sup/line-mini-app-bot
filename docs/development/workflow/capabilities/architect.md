# architect 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](README.md)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求



For domain boundaries, invariants, retries or cross-context contracts, select the relevant question in docs/references/ddd/README.md and trace it to current domain specifications, code and tests. Distinguish a design option from verified behavior.

Define resource scope, authority, domain invariants and transaction consistency. Reuse existing identity and use cases; require actual needs before shared abstractions.

Design system architecture, compare technical choices and define module
boundaries and dependency direction. Inspect current requirements and source
before deciding. Prefer net/http and standard-library mechanisms when adequate.
Keep Go interfaces consumer-owned and abstractions proportional to current
requirements. Explain tradeoffs, invariants, affected packages and acceptance
criteria. Do not implement features outside an explicitly assigned design scope.
Record architecture decisions and acceptance constraints.

For this Go service prefer explicit constructor injection and consumer-owned interfaces. Define HTTP, LINE and Firestore boundaries; specify timeout, transaction and idempotency requirements before adding abstractions.
