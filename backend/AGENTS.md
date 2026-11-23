<a id="backend-工作指引"></a>

# backend agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [package.json](package.json), [turbo.json](turbo.json), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Documentation index](../docs/README.md) · [Module entry points](../docs/README.md#按模組閱讀)

Applies to this directory and its descendants. Follow the [root AGENTS.md](../AGENTS.md), then nearer directory guidance. User and host instructions take precedence; this file grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

The sole executable entry point is [main.go](main.go). [business](internal/business) is organized by domain; [bootstrap](internal/bootstrap) composes services; [identity](internal/identity) handles identity/sessions; [httpapi](internal/transport/http/api) owns OpenAPI wiring and generated code.

Separate business cores, HTTP/LINE adapters and Firestore adapters; shared composition lives in internal/bootstrap. Follow ../api contract workflows before changing HTTP formats. Do not edit generated code or put credentials in .env.example.

## Domain & Reliability Gate

Applies to Go backend additions, changes and refactors. Follow domain guidance and the verification workflow below.

<a id="業務與架構"></a>

### Business and architecture

- MUST clarify business goals, bounded context, state, data relationships and existing call chains first. Domains own their data/state and collaborate through explicit contracts without dependency cycles.
- MUST keep business rules, authoritative calculations and transitions in their domains. Use cases coordinate authorization and consistency. Obtain actors from trusted sessions or verified LINE identities; decide by resource, action, policy and current state. MUST NOT trust client-declared roles or permissions.
- MUST preserve dependency direction across Domain, Use Case, Port, Adapter and Bootstrap. Keep HTTP DTOs, Firestore formats, LINE SDK and framework types outside the core. Handlers/adapters must not recalculate domain results or invent authorization rules.
- MUST define transaction boundaries, idempotency and failure semantics for concurrency, retries and partial success. Before changing public behavior or persistence, check consumers, historical document read/write compatibility and migration risks.

<a id="最小化與驗收"></a>

### Minimal changes and acceptance

- SHOULD prefer removing, moving, merging or reusing Go code. MUST NOT add interfaces, services, factories, managers, registries or generic engines solely for one implementation, hypothetical needs or uniformity, or merge distinct domain rules for reuse.
- MUST trace callers from the original symptom and fix root causes together. For HTTP formats, change contracts and regenerate through the [API workflow](../api/README.md).
- MUST verify success, rejection, failure and key boundaries as specified below. For Firestore transactions, concurrency, retries or persistence compatibility, run [existing emulator integration acceptance](../docs/development/tooling/testing/commands.md#firestore-integration).

**Completion criteria**: one business authority, correct dependencies, verified authorization/consistency and passing applicable checks. Continue fixing acceptance failures; identify unverifiable cases when blocked.

<a id="驗證"></a>

## Verification

Check callers and affected Go tests when changing package behavior.

Use the root scripts/tooling/tool-exec.sh. Follow the [acceptance matrix](../docs/development/tooling/testing/commands.md#變更驗收矩陣) for fmt-check, lint, check, build and relevant existing tests after Go changes. Do not claim application tests for documentation-only work.

Run commands from the repository root unless stated otherwise. Read the [local README](README.md) and [documentation index](../docs/README.md); report actual results and limitations.

<a id="文件閱讀路由"></a>

## Documentation routing

Choose sources through [related module documentation](README.md#相關文件) and [module entry points](../docs/README.md#按模組閱讀). Maintain local purposes and commands in this module's README; organize cross-module architecture, feature and operations documents by topic under docs. Follow the [documentation maintenance guidance](../docs/AGENTS.md#主題子目錄) when updating indexes.

## Business Operation

Before related work, read the [operation model](../docs/contracts/business-operations.md) and [compliance workflow](../docs/development/workflow/implementation.md#business-operation-compliance).

The backend obtains actors from trusted identities and checks Authority/Policy by resource ownership and scope. Use cases coordinate rules/results. Coordinate consistency-sensitive authorization, version and state checks with transactions; define retry and side-effect failure semantics. Do not create generic engines or access modules without actual need.

[Current implementation diagram: formats and data compatibility](../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="前端操作與業務契約"></a>

## Frontend operations and business contracts

Review operation results using [frontend, API and backend responsibilities and retry examples](../docs/contracts/business-operations.md#前端操作與業務契約).

Before adding, editing or reviewing business models, API contracts, resource paths, collections or query fields, all roles must read the [naming standards](../docs/standards/naming.md#業務實體集合與-api-命名). Use singular names for individual business entities and plural names for collections and REST resource paths, respecting existing compatibility exceptions. Maintain the full rules only in that document; do not directly rename existing public contracts or persisted data.

The backend executes authorization/business rules; domains retain independent state models and share necessary mechanisms without unnecessary abstraction. Follow the [geolocation specification](../docs/integrations/geolocation/browser-api.md) for location authorization and validation.

## Applicable engineering conventions

- **Formatting and types:** Use pinned Go formatting/lint/type checks; keep domain values distinct from generated HTTP and persistence representations.
- **Consistency:** Define the transaction, version condition and idempotency boundary for each write; retry external calls only with bounded deadlines and explicit recovery semantics.
- **Diagnostics:** Use structured errors/logs and correlated operation identifiers; keep secrets out of logs and metric labels bounded.
- **Acceptance:** Choose domain, contract, emulator and end-to-end checks by the changed boundary; publish only verified, independently reviewable milestones.
