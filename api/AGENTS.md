<a id="api-工作指引"></a>

# api agent guidance

See the [API workflow](./) for directory responsibilities, generation and verification commands. Read the relevant guidance before editing [contract](contract/AGENTS.md), [codegen](codegen/AGENTS.md) or [client](client/AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Business entity, collection and API naming](../docs/standards/naming.md#業務實體集合與-api-命名) · [Documentation index](../docs/)

Applies to this directory and its descendants. Follow the [root AGENTS.md](../AGENTS.md), then nearer directory guidance. User and host instructions take precedence; this file grants no additional authorization for external writes.

<a id="巢狀指引"></a>

## Nested guidance

Read the relevant guidance before entering [contract](contract/AGENTS.md), [client](client/AGENTS.md) or [codegen](codegen/AGENTS.md). Capability directories under contract also inherit contract/AGENTS.md. Each level adds only its own responsibilities without duplicating parent rules. See the [directory guide](./).

<a id="職責與邊界"></a>

## Responsibilities and boundaries

Before changing contracts, read [business entity, collection and API naming](../docs/standards/naming.md#業務實體集合與-api-命名). Name model types, collections, queries and paths by their semantics. Preserve existing singleton and public-contract compatibility exceptions; do not manually change generated names.

contract/openapi.yaml only composes public contracts. Follow the [contract capability index](contract/) for ownership and references. Each capability owns its operations and DTOs; shared errors and authentication belong in common/. Capability files must not reference the root entry point. Keep client/ and codegen/ shared rather than duplicating them per capability.

Change contracts first, then run pnpm api:generate from the root to update Go/TS output. Do not edit generated files manually. The business core must not depend on HTTP DTOs.

## Contract & Compatibility Gate

Applies to public and internal HTTP contracts, DTOs, clients and generation. Follow nested contract, client and codegen guidance.

<a id="契約與相容性"></a>

### Contracts and compatibility

- MUST use OpenAPI as the sole authority for HTTP contracts. First check capability ownership, resource semantics, data ownership and consumers in [capabilities.json](contract/capabilities.json), distinguishing public HTTP, internal worker HTTP and non-HTTP events.
- MUST define requests, responses, status codes, authentication, errors and necessary retry semantics using shared error and authentication definitions. APIs describe observable behavior; the backend executes business rules, state transitions and authorization. Client-supplied actors, roles, permissions or decisions are not trusted authorization evidence.
- MUST preserve compatibility of public paths, fields, types, errors and behavior. Before breaking changes, confirm requirements, affected callers and migration strategy.

<a id="最小化與驗收"></a>

### Minimal changes and acceptance

- SHOULD extend operations with matching semantics and return consumer-required data to reduce transfer and repeated assembly. MUST NOT break resource semantics or other consumers' compatibility for one screen, or force schemas with different semantics to share definitions.
- MUST NOT stack DTOs, wrappers, schemas or converters for one field or hypothetical needs. Reuse generation settings and shared clients; avoid duplicate endpoints, parallel contracts and extra generators.
- MUST confirm backend business behavior first, then synchronize Contract → Generate → Backend Adapter → Frontend Consumer. Review Go/TypeScript output and all affected consumers. Do not edit generated files or duplicate transport types.
- MUST run `pnpm api:check` using the verification workflow below to check operation ownership, schema references and generated output, excluding cycles, invalid definitions and contract drift. Compare historical compatibility against an explicitly specified base revision. Verify success, rejection, errors and applicable retry scenarios for changed operations.

**Completion criteria**: explicit operation semantics, consistent contracts and implementations at both ends, no duplicate definitions or invalid endpoints, and passing applicable compatibility checks. Report verification gaps; do not weaken contracts or change comparison baselines to pass acceptance.

<a id="驗證"></a>

## Verification

Run bash scripts/tooling/tool-exec.sh pnpm api:check from the repository root. State the base revision and results for historical compatibility comparisons. See the [API workflow](./) for check coverage.

Run commands from the repository root unless stated otherwise. Read the [local README](./) and [documentation index](../docs/); report actual results and limitations.

<a id="文件閱讀路由"></a>

## Documentation routing

Choose sources through [related module documentation](./) and [module entry points](../docs/). Maintain local purposes and commands in this module's README; organize cross-module architecture, feature and operations documents by topic under docs. Follow the [documentation maintenance guidance](../docs/AGENTS.md#主題子目錄) when updating indexes.

## Business Operation

Before related work, read the [operation model](../docs/contracts/business-operations.md) and [compliance workflow](../docs/development/workflow/implementation.md#business-operation-compliance).

Operation contracts must map to OpenAPI inputs, responses and existing public failure semantics. Clients cannot declare trusted actors, roles or review authority. Change contracts before generation; do not edit generated code, force new endpoints or break compatibility to fit the model.

[Current implementation diagram: basic entry points](../docs/contracts/http.md#基礎入口)

<a id="前端操作與業務契約"></a>

## Frontend operations and business contracts

Review operation results using [frontend, API and backend responsibilities and retry examples](../docs/contracts/business-operations.md#前端操作與業務契約).

Before adding, editing or reviewing business models, API contracts, resource paths, collections or query fields, all roles must read the [naming standards](../docs/standards/naming.md#業務實體集合與-api-命名). Use singular names for individual business entities and plural names for collections and REST resource paths, respecting existing compatibility exceptions. Maintain the full rules only in that document; do not directly rename existing public contracts or persisted data.

## Applicable engineering conventions

- **API and versioning:** Name operations by resource semantics and preserve existing public paths, status codes and error meanings; document a migration before incompatible changes.
- **Generation and types:** Update the owning schema, capability consumers and generated Go/TypeScript artifacts together; verify the contract after frontend calls are added, moved or removed.
- **Validation:** Define missing/null/zero values and list pagination, filters, stable ordering and projections in the contract rather than in screen-specific DTOs.
