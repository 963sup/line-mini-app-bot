# Backend module rules

This file adds backend-specific rules to the [repository rules](../AGENTS.md). Directory purposes belong in the [root README](../README.md#巢狀目錄導覽). Detailed path constraints are indexed in [module constraints](../docs/standards/module-constraints.md).

## Hexagonal architecture

Business contexts live under `internal/business/<context>/` and adopt only the layers they need:

```text
domain/{entity,valueobject,service,event,policy}
application/{command,query,port,dto}
adapter/{http,line,firestore,messaging,scheduler}
bootstrap
```

- **Dependency direction.** Dependencies MUST point inward. Domain code MUST remain independent of HTTP, LINE, Firestore and framework types. Applications coordinate domain rules and define output ports. Adapters call use cases or implement ports. `bootstrap/` composes validated dependencies.
- **Minimal structure.** A context MUST use only packages with a real responsibility. Code MUST NOT be moved or empty layers created merely to match the reference tree.
- **Business authority.** Resource-owning domains MUST define authorization, calculations, state and legal transitions. Actors MUST come from trusted sessions or verified provider identities. Adapters MUST NOT invent authorization or recalculate domain results.
- **Consistency.** Consistency-sensitive authorization, version checks and state transitions MUST share the required transaction boundary. External effects MUST stay outside retried storage callbacks. Writes MUST define idempotency, partial-failure and recovery semantics.
- **Data compatibility.** Resource IDs, stored encodings and historical readers MUST remain compatible during model or serialization changes unless a migration explicitly handles them.
- **Contract order.** Public HTTP formats MUST change in OpenAPI first and generated files MUST NOT be edited manually.
- **Reliability.** External calls MUST have bounded deadlines, cancellation and typed failure handling. Secrets and sensitive payloads MUST stay out of logs; metric labels MUST remain bounded.
- **Verification.** Go changes MUST run the checks selected by the [acceptance matrix](../docs/development/tooling/testing/commands.md#變更驗收矩陣). Firestore transaction, retry or persistence changes MUST run the emulator integration workflow.

Current package responsibilities and allowed dependencies are authoritative in [backend boundaries](../docs/architecture/runtime/backend-boundaries.md).
