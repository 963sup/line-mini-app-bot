# Backend module rules

This file adds backend-specific rules to the [repository rules](../AGENTS.md). Directory purposes belong in the [root README](../README.md).

- **gopls MCP.** Before editing existing Go symbols, contributors MUST use `gopls` to inspect applicable definitions, references and package APIs. After editing Go code, contributors MUST request `gopls` diagnostics for the changed packages before running the required format, test and architecture checks. Direct source search MAY supplement but MUST NOT replace symbol-aware analysis when `gopls` is available. If the `gopls` MCP server is unavailable, contributors MUST report the limitation and use `go doc`, `go list`, targeted source search and Go checks as the fallback.

## Hexagonal architecture

Business contexts live under `internal/business/<context>/` and adopt only the layers they need:

```text
domain/{entity,valueobject,service,event,policy}
application/{command,query,port,dto}
adapter/{http,line,firestore,messaging,scheduler}
bootstrap
```

- **Dependency direction.** Dependencies MUST point inward. Domain code MUST remain independent of HTTP, LINE, Firestore and framework types. Application code MUST coordinate use cases and their transaction requirements. Adapters MUST translate external formats and either call use cases or implement their ports. `bootstrap/` composes validated dependencies.
- **Domain model.** Each bounded context MUST own its invariants and legal state transitions. Entities, value objects and aggregates SHOULD be introduced where identity, value semantics or an atomic invariant requires them. A context MUST NOT add empty layers to match the reference tree.
- **Ports and injection.** Go interfaces at a capability boundary MUST be defined by the caller and expose only the methods that caller needs. `bootstrap/` MUST inject adapters through explicit constructors or configuration. A new interface, mapper or forwarding Service MUST have a distinct boundary responsibility.
- **Business authority.** Resource-owning domains MUST define authorization and calculations. Use cases MUST obtain actors from trusted sessions or verified provider identities, check the requested action against the resource scope and context, and deny it unless allowed by policy. Writes MUST enforce domain invariants before changing state. Adapters MUST NOT invent authorization or recalculate domain results.
- **Commands, queries and events.** Queries MUST NOT request a business state transition. Commands MUST define preconditions and a result. When emitted, a domain event MUST represent an occurred business fact with an identity, resource reference and sufficient trace context for the documented delivery behavior. External publication MUST follow the successful state change.
- **Consistency.** Consistency-sensitive authorization, version checks and state transitions MUST share the required transaction boundary. External effects MUST stay outside retried storage callbacks. Writes MUST define idempotency, partial-failure and recovery semantics.
- **Data compatibility.** Resource IDs, stored encodings and historical readers MUST remain compatible during model or serialization changes unless a migration explicitly handles them.
- **Contract order.** Public HTTP formats MUST change in OpenAPI first and generated files MUST NOT be edited manually.
- **Go conventions.** Go code SHOULD use standard library primitives where they meet the use case and follow the [backend standard](../docs/standards/backend.md). I/O MUST propagate the caller's `context.Context`. Outbound adapters MUST bound calls and translate SDK errors at their boundary while preserving cancellation and the caller's documented error semantics.
- **Observability.** Secrets and sensitive payloads MUST stay out of logs. Metric labels MUST remain bounded.

## Verification

For LINE changes, use the [LINE ownership map](../docs/architecture/line.md) and [LINE interaction standard](../docs/standards/line-interactions.md). Add outbound Messaging API operations under `internal/platform/line`, signature verification and dispatch under `internal/transport/line`, and feature-specific event translation under that context's `adapter/line`. Read deployment settings and inject clients in bootstrap. Business core must use its own types. Keep current delivery deadlines, retry keys, webhook ordering and historical Rich Menu aliases covered by tests.

- Go changes MUST run the checks selected by the [verification workflow](../docs/workflows/verification.md). Changes to module dependencies MUST run the module-root [architecture boundary test](architecture_test.go) and its [regression test](architecture_regression_test.go). Changes to policy or transitions MUST verify applicable allowed and denied actions and illegal transitions, including a resource outside the authorized scope and unchanged persisted state after a rejected write where applicable. Firestore transaction, retry or persistence changes MUST run the emulator integration workflow.

Current package responsibilities and allowed dependencies are authoritative in [architecture boundaries](../docs/architecture/boundaries.md).
