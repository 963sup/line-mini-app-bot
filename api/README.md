<a id="http-api-契約"></a>

# HTTP API contracts

Maintains API contracts, generation settings and the TypeScript client. Change contracts before regenerating Go and TypeScript files whenever API inputs or outputs change.

Contracts define interfaces; consumers use the client and generation settings to regenerate contract types.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

For endpoints or data formats, select the owning domain in the [contract capability index](contract/README.md). For generators, read the [generation settings guide](codegen/README.md); for TypeScript clients, read the [client guide](client/README.md). Read the [local agent guidance](AGENTS.md) before editing.

<a id="相關規格"></a>

## Related specifications

[Independent projects](../docs/domains/project/project.md) · [React frontend lifecycle](../docs/architecture/runtime/react-lifecycle.md)

Engineering groups and projects have independent entry points. Existing repository-scoped project routes use a compatibility adapter calling the same use cases; source identity no longer defines project ownership. See [independent projects](../docs/domains/project/project.md) for data, authorization and compatibility boundaries. Follow the linked generation workflow; do not edit generated files.

[contract/openapi.yaml](contract/openapi.yaml) is the sole composition entry point for public HTTP contracts. Capabilities own their operations and formats; shared errors and authentication live in `contract/common/`. See the [contract capability index](contract/README.md) for ownership. Do not edit generated files or import generated packages into the business core.

See the [contract generation workflow](../docs/development/tooling/api/change-api.md#契約生成流程) for generation and check commands.

See the [React frontend lifecycle](../docs/architecture/runtime/react-lifecycle.md) for browser controllers and SessionProvider authentication requests.

<a id="相容性與-ci"></a>

## Compatibility and CI

api:check verifies synchronization of current OpenAPI contracts, routes and generated code.

Configuration, identity, session, repository and attendance operations (including leave registration) are defined by `contract/openapi.yaml` and its references; do not duplicate counts here. This covers our MINI App HTTP API. LINE webhooks, internal summary jobs and the independent geolocation example are outside this public contract; do not duplicate official LINE APIs.

<a id="保持的資料語意"></a>

## Preserved data semantics

- Worktime includes attendance, estimated totals and confirmation status. Existing Minutes remains confirmed net worktime. EstimatedMinutes already includes Minutes; do not add them together. See [worktime calculation](../docs/domains/attendance/worktime.md).
- Preserve existing casing: shift responses use RecordID/ClockInAt; correction inputs use recordId/start.
- Define and preserve null separately from empty arrays. Times use RFC3339; months use Asia/Taipei. The attendance core remains responsible for cross-month calculation and complete summaries.
- Coordinates use double to avoid precision changes from generated Go float32 values. Worktime Duration retains its existing integer nanosecond format; Minutes contains calculated minutes.
- Detail lists use opaque cursors, defaulting to 20 items with a maximum of 100. Monthly summaries are independent and unaffected by pagination. Empty pageSize/cursor values retain default semantics.
- New clients should retain an Idempotency-Key for each correction draft and reuse it on retries. The same user/key/content returns the original result; different content returns 409. Legacy clients may omit keys but have no cross-request deduplication guarantee.
- Retain existing non-paginated groups and monthly endpoints as deprecated; do not add compatibility routes.

<a id="驗證邊界"></a>

## Validation boundaries

Attendance requests pass CSRF and authentication (cookie sessions or existing cookieless bearer credentials), then OpenAPI structural validation, then existing authorization and use cases. Cache verified identity only in the request context to avoid repeated identity-service calls within a request. Cross-request session state is stored separately in Firestore. Retain 503 when attendance is unconfigured. HEAD follows GET semantics.

OpenAPI validation does not replace shift versions, group membership, review permissions, location freshness or atomic transactions. Validator errors must not expose locations, application content or tokens. Preserve existing request limits of 512/2048/16384 bytes.

Validate requests at runtime. Use tests for response contracts and JSON consistency between old and new DTOs rather than buffering every production response for validation. Standard CSRF rejection remains text/plain 403 and is documented in write endpoint contracts.

<a id="後端實作位置"></a>

## Backend implementation locations

Generated code and shared wiring remain in backend/internal/transport/http/api. Attendance handlers, DTO mapping and runtime validation live in business/attendance/adapter/http; identity entry points live in identity/http. Core types have no JSON/Firestore tags; storage adapters encode documents and translate errors.

Redocly lint/bundle settings live in `api/codegen/redocly.yaml`. Root generation scripts explicitly pass `--config` instead of relying on working-directory discovery.

<a id="目錄導覽與維護"></a>

## Directory guide and maintenance

[Contract sources](contract/README.md) are organized by capability; the [client](client/README.md) and [generation settings](codegen/README.md) are shared. Validate contract changes through the [generation workflow](../docs/development/tooling/api/change-api.md#契約生成流程). For historical compatibility comparisons, state revisions and differences; current `api:check` checks contracts and generated output in the working tree.

[Local guidance](AGENTS.md) · [Root README](../README.md) · [Root guidance](../AGENTS.md)

<a id="巢狀指引"></a>

## Nested guidance

- [Contract sources](contract/AGENTS.md): see the [contract index](contract/README.md#巢狀指引) for capability guidance.
- [Shared client](client/AGENTS.md).
- [Generation settings](codegen/AGENTS.md).

Basic Issue operations, Discussion topics/comments and Repository label catalogs are wired. Project/items and Repository milestones/statistics follow their basic specifications; Project fields/views remain scaffolds.

<a id="相關文件"></a>

## Related documentation

- [Documentation index: API](../docs/README.md#按模組閱讀)
- [Sessions and sliding renewal](../docs/contracts/identity.md)
- [HTTP contract design](../docs/contracts/http.md)
- [Data contracts](../docs/proposals/firestore-data-contracts.md)
- [Attendance queries and corrections](../docs/domains/attendance/queries.md)

- [Group leave registration](../docs/domains/attendance/leave.md)

- [Current modules and responsibilities](../docs/architecture/runtime/code-layout.md)

<a id="business-入口分類"></a>

## Business entry point classification

Classify public HTTP, events and internal jobs through the [contract capability index](contract/README.md). Summary and Weather events and internal jobs retain their existing entry points.

- [Group repositories](../docs/domains/repository/README.md)
- [Space label catalog](../docs/domains/repository/labels.md): list/create catalog entries; classify through Issue/Discussion updates. No rename or deletion.
- [Space issues, phase one](../docs/domains/repository/issue/design.md): public Issue contracts and generated types are wired; production acceptance remains pending.
- [Space discussions and comments](../docs/domains/repository/discussion/design.md): topic and single-level comment contracts with author editing.

- [Repository, Issue and Discussion models and Connection pagination proposal](../docs/proposals/platform/domain-model.md): future design reference; current OpenAPI contracts remain unchanged.

- [Milestones](../docs/domains/repository/milestones.md), [Project](../docs/domains/project/project.md) and [Insights](../docs/domains/repository/insights.md): formal contracts and wiring across all three layers.

<a id="協作能力分佈與交付狀態"></a>

## Collaboration capability coverage and delivery status

See the [capability inventory](contract/capabilities.json) for wired and planned operations, and the [OpenAPI contract](contract/openapi.yaml) for requests and responses.

[Current implementation diagram: basic entry points](../docs/contracts/http.md#基礎入口)
