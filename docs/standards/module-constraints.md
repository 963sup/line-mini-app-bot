# 模組必要約束

本文件保存已從巢狀 `AGENTS.md` 收斂的模組專屬限制。根目錄與頂層模組指引負責執行規則；一般目錄用途與導航由根目錄 README 維護；業務、契約與架構細節仍以各節連結的權威文件及目前實作為準。

只有修改對應路徑時才讀取該節。規則句使用「MUST／MUST NOT／SHOULD」表示必要、禁止及有理由才可偏離。

## `api/client/`

- **Types and generation.** MUST export generated request/response types through the shared client; repair generation sources instead of adding handwritten parallel types.
- **Identity and retries.** MUST keep credential, cancellation and retry behavior explicit per client; preserve the original request identity when an outcome is unknown.
- **Compatibility.** MUST check existing client callers when changing defaults, serialization or error propagation; never bundle internal worker credentials.

## `api/client/src/`

- **Type safety.** MUST map client calls to generated operation signatures and preserve optional, nullable and error fields through wrappers.
- **Serialization and errors.** MUST forward contract-defined dates, identifiers and recovery metadata without coercing them into a second wire format.
- **Verification.** MUST check affected consumers and generated-output freshness after changing client exports or request defaults.

## `api/codegen/`

- **Reproducible generation.** MUST use pinned generators and the existing entry points to rebuild artifacts from schemas; review output drift before committing.
- **Structure and dependencies.** MUST keep public/internal generation separate and generated transport types outside domain packages.
- **Compatibility.** MUST preserve operation names, field casing and identifier mappings unless the source contract explicitly changes them.

## `api/contract/`

- **Schema and ownership.** MUST keep one owning capability for each operation and DTO, with shared transport definitions referenced rather than copied.
- **HTTP and querying.** MUST specify request validation, stable errors, pagination scope and ordering before wiring an operation or registering consumers.
- **Evolution.** MUST synchronize affected consumers and generated artifacts after schema or call-site changes, then run the existing API contract checks; retain compatible readers during expand-contract migrations.
- **Types and compatibility.** MUST preserve field casing, optionality and existing operation/error semantics; make incompatible schema evolution an explicit migration.

## `api/contract/attendance/`

- **Contract semantics.** MUST keep instant formats, business dates, review authority and original-request retry semantics explicit for attendance operations.

## `api/contract/common/`

- **Contract semantics.** MUST keep authentication schemes and the shared error envelope transport-only; use stable error codes and optional recovery fields without duplicating domain schemas.

## `api/contract/discussion/`

- **Contract semantics.** MUST specify topic/comment lifecycles and edit versus classification permissions independently; never equate answer selection with business approval.

## `api/contract/group/`

- **Contract semantics.** MUST distinguish a group identifier, observed interaction candidates and verified membership in every visible group response.

## `api/contract/identity/`

- **Contract semantics.** MUST specify token/session expiry, CSRF requirements and revocation errors without implying that authentication grants business authorization.

## `api/contract/issue/`

- **Contract semantics.** MUST keep Issue transitions and versions separate from comment, label and milestone references; define concurrent edit and retry outcomes.

## `api/contract/journal/`

- **Contract semantics.** MUST specify published journal visibility, source redaction, work-date representation and idempotent publication results together.

## `api/contract/object/`

- **Contract semantics.** MUST preserve exact quantity representations, object/parent identifiers and version conditions across object, subobject and history schemas.

## `api/contract/project/`

- **Contract semantics.** MUST use the owning Project contracts for Owner types and source requirements; distinguish item IDs, source IDs and field-option IDs.

## `api/contract/repository/`

- **Contract semantics.** MUST scope repository and child-resource queries explicitly; returning a repository or capability flag does not authorize a later write.

## `api/contract/summary/`

- **Contract semantics.** MUST keep internal worker identity, event/job identifiers, lease expiry and duplicate delivery semantics separate from browser sessions.

## `api/contract/weather/`

- **Contract semantics.** MUST describe the existing event input and forecast validity without introducing a public endpoint or expanding location-data retention.
- **Event compatibility.** MUST follow existing provider/event schemas and receiver version support; do not add HTTP generation or a CloudEvents envelope merely for uniformity.

## `backend/internal/`

- **Structure and naming.** MUST place each change in its existing bounded context and responsibility package; do not create empty layers to match a template.
- **Types and dependencies.** MUST keep domain code independent of transport/storage SDKs and inject ports at composition entry points.
- **Data evolution.** MUST preserve resource IDs, stored encodings and old-reader compatibility when moving models or changing serialization.

## `backend/internal/architecture/`

- **Dependency enforcement.** MUST express architectural boundaries through actual package imports and focused tests, without exemptions that hide new cycles.
- **Structure.** MUST verify callers and allowed dependency direction after package moves; a matching directory name alone is not evidence of a valid boundary.
- **Documentation.** MUST keep boundary tests, architecture diagrams and accepted decisions consistent with the same implementation revision.

## `backend/internal/bootstrap/`

- **Dependency injection and configuration.** MUST assemble existing ports/adapters from validated settings; do not move business decisions into constructors.
- **Lifecycle and reliability.** MUST give clients and background work explicit cancellation, deadlines and cleanup ownership, including partial startup failure.
- **Observability and security.** MUST when adding instrumentation, prefer OpenTelemetry-compatible traces/metrics and correlate structured logs; keep credentials out and metric label cardinality bounded.

## `backend/internal/business/`

- **Domain boundaries.** MUST use each context's ubiquitous language and owned interfaces; shared mechanics must not merge independent business states or permissions.
- **State and transactions.** MUST coordinate authorization, version checks and legal transitions at the required transaction boundary; keep external effects outside retried storage callbacks.
- **Events and lifecycle.** MUST for emitted events or jobs, define identity, schema version, duplicate handling and cleanup responsibility before adding consumers.
- **Compatibility and tests.** MUST preserve public errors and stored formats; verify changed rules with focused domain tests and applicable contract/emulator cases.

## `backend/internal/business/agent/`

- **Domain, state and authorization.** MUST validate model-proposed operations against the task allowlist, resource scope and step budget before executing side effects.
- **Reliability and lifecycle.** MUST bound model/tool calls, preserve cancellation and stop at the declared task budget; retries must not expand the operation allowlist.
- **Verification.** MUST for implemented behavior, test denied operations, exhausted evidence/budgets and termination outcomes without inventing persistent workflow infrastructure.

## `backend/internal/business/attendance/`

When changing this module, you MUST check the [decision flow](../domains/attendance/design.md) and [representative tests](../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go).

- **Domain, state and authorization.** MUST use the attendance contract's business time zone and domain transitions for punches, corrections and reviews; adapters must not independently recompute worktime.

## `backend/internal/business/discussion/`

When changing this module, you MUST check the [decision flow](../domains/repository/discussion/design.md) and [representative tests](../../backend/internal/business/discussion/domain/entity/classification_test.go).

- **Domain, state and authorization.** MUST check content-edit and classification authority separately, and preserve comment/topic versions and lifecycle semantics on retries.

## `backend/internal/business/enterprise/`

When changing this module, you MUST check the [decision flow](../domains/enterprise/design.md) and [implementation](../../backend/internal/business/enterprise/adapter/firestore/store.go).

- **Domain, state and authorization.** MUST keep provisioning eligibility, enterprise ownership and team-management authority distinct, including retries returning an already-created enterprise.

## `backend/internal/business/issue/`

When changing this module, you MUST check the [decision flow](../domains/repository/issue/design.md) and [representative tests](../../backend/internal/business/issue/domain/entity/issue_test.go).

- **Domain, state and authorization.** MUST apply Issue-owned transition and edit policies without importing attendance approval states or granting content edits through classification.

## `backend/internal/business/journal/`

When changing this module, you MUST check the [decision flow](../domains/journal/journal.md) and [representative tests](../../backend/internal/business/journal/domain/entity/entry_test.go).

- **Domain, state and authorization.** MUST retain published content visibility while redacting unauthorized source data; preserve historical payload and receipt-hash encodings during model changes.

## `backend/internal/business/knowledge/`

- **Domain, state and authorization.** MUST for implemented publication/read paths, authorize scope and version visibility explicitly; drafts or withdrawn knowledge must not become silently available to new tasks.
- **Version and cache lifecycle.** MUST for implemented readers, scope cached knowledge by authorized version and invalidate it when publication or withdrawal changes visibility.
- **Acceptance.** MUST keep proposal-only policies separate from deployed behavior and test cross-organization denial whenever a reader is implemented.

## `backend/internal/business/object/`

When changing this module, you MUST check the [decision flow](../domains/repository/object/design.md) and [representative tests](../../backend/internal/business/object/application/progress_test.go).

- **Domain, state and authorization.** MUST validate exact quantities, parent scope and legal progress transitions in the domain; record history only for committed, authorized version changes.

## `backend/internal/business/operations/`

- **Domain, state and authorization.** MUST keep extraction/model adapters focused on I/O and translate their failures into application errors; the resource owner retains business authorization.
- **Errors and reliability.** MUST bound document/model input, propagate cancellation and map provider/process failures to stable application errors rather than leaking SDK-specific failures.
- **Testing and privacy.** MUST verify malformed, oversized, cancelled and unavailable extraction paths; do not retain source documents or add writes implicitly.

## `backend/internal/business/organization/`

When changing this module, you MUST check the [decision flow](../domains/organization/design.md) and [implementation](../../backend/internal/business/organization/adapter/firestore/store.go).

- **Domain, state and authorization.** MUST evaluate ownership and membership together with transfers/removals; legacy document translation must not grant rights from unrelated owner fields.

## `backend/internal/business/project/`

When changing this module, you MUST check the [decision flow](../domains/project/project.md) and [representative tests](../../backend/internal/business/project/application/reference_source_test.go).

- **Domain, state and authorization.** MUST authorize the Project Owner and each referenced source independently; copying creates a new resource identity without inheriting unauthorized source access.

## `backend/internal/business/repository/`

When changing this module, you MUST check the [decision flow](../domains/repository/design.md) and [implementation](../../backend/internal/business/repository/domain/entity/repository.go).

- **Domain, state and authorization.** MUST revalidate repository ownership and active status at restricted writes; list filters and cached capability flags are not write authorization.

## `backend/internal/business/summary/`

When changing this module, you MUST check the [decision flow](../domains/summary/summary.md) and [implementation](../../backend/internal/business/summary/application/summarization/validation.go).

- **Domain, state and authorization.** MUST preserve source/citation identity, bounded model input and job expiry across retries; expired or inaccessible messages must not enter a new result.
- **Concurrency and recovery.** MUST coordinate job leases, completion and deduplication at the existing storage boundary; repeated delivery must not duplicate publication.
- **Data compatibility.** MUST preserve stored receipt/input encodings and source expiry during model changes; verify citation and lease-recovery boundaries.

## `backend/internal/business/team/`

When changing this module, you MUST check the [decision flow](../domains/team/design.md) and [implementation](../../backend/internal/business/team/adapter/firestore/store.go).

- **Domain, state and authorization.** MUST keep team-side management checks distinct from destination-organization authority and reject conflicting management assignments.

## `backend/internal/business/weather/`

When changing this module, you MUST check the [decision flow](../domains/weather/weather.md) and [representative tests](../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go).

- **Domain, state and authorization.** MUST treat location inputs as purpose-limited data and make forecast validity, cache expiry and unavailable-result behavior explicit.
- **Cache and reliability.** MUST scope cached forecasts by their documented location/time semantics, bound provider calls and distinguish expired, unavailable and valid results.
- **Event acceptance.** MUST test replyable group events, invalid input and provider failure without introducing location history or attendance proof.

## `backend/internal/business/workflow/`

- **Domain, state and authorization.** MUST for implemented workflows, persist the resumable step and operation identity; interruption or repeated confirmation must not duplicate completed effects.
- **State and idempotency.** MUST define legal pause/resume/complete transitions and reuse the original write identity when resuming implemented steps.
- **Acceptance.** MUST keep proposed orchestration separate from deployed code and verify interruption before confirmation and after a successful side effect.

## `backend/internal/identity/`

- **Authentication and security.** MUST keep trusted actor identifiers distinct from profiles and resource ownership; identity verification does not assign domain permissions.
- **Time and identifiers.** MUST preserve token/session identifier formats and expiry precision; distinguish UTC instants from local business dates.
- **Reliability and tests.** MUST bound verification calls and test invalid, expired, revoked and cancelled cases at the changed boundary.

## `backend/internal/identity/google/`

- **Authentication and security.** MUST verify provider token claims against the configured trust requirements; reject expired or mismatched credentials without logging token contents.
- **Time and identifiers.** MUST preserve token/session identifier formats and expiry precision; distinguish UTC instants from local business dates.
- **Reliability and tests.** MUST bound verification calls and test invalid, expired, revoked and cancelled cases at the changed boundary.

## `backend/internal/identity/http/`

- **Authentication and security.** MUST preserve Cookie/Bearer and CSRF semantics, stable authentication errors and no-store behavior for private session responses.
- **Time and identifiers.** MUST preserve token/session identifier formats and expiry precision; distinguish UTC instants from local business dates.
- **Reliability and tests.** MUST bound verification calls and test invalid, expired, revoked and cancelled cases at the changed boundary.

## `backend/internal/identity/line/`

- **Authentication and security.** MUST verify LINE credentials through the existing provider boundary; do not accept browser-supplied profile IDs as authenticated actors.
- **Time and identifiers.** MUST preserve token/session identifier formats and expiry precision; distinguish UTC instants from local business dates.
- **Reliability and tests.** MUST bound verification calls and test invalid, expired, revoked and cancelled cases at the changed boundary.

## `backend/internal/identity/session/`

- **Authentication and security.** MUST enforce session expiry, renewal and revocation using the authoritative record and required concurrency checks; stale cached state must not revive a revoked session.
- **Time and identifiers.** MUST preserve token/session identifier formats and expiry precision; distinguish UTC instants from local business dates.
- **Reliability and tests.** MUST bound verification calls and test invalid, expired, revoked and cancelled cases at the changed boundary.

## `backend/internal/linegroup/`

When changing this module, you MUST check the [decision flow](../integrations/line/group-compatibility.md#decision-flow) and [representative tests](../../backend/internal/linegroup/application/query/group_test.go).

- **Authorization.** MUST resolve restricted group access through verified membership; cached candidates and webhook observation do not prove current membership.
- **Caching and queries.** MUST scope membership/cache keys by the relevant identity and group, with explicit expiry and failure behavior.
- **Compatibility.** MUST preserve legacy group identifiers while testing that unavailable or changed membership cannot reuse stale authorization.

## `backend/internal/platform/`

- **Dependencies.** MUST keep platform helpers limited to shared mechanisms and inject provider clients rather than embedding business policy.
- **Reliability.** MUST propagate cancellation and typed failures, bound call deadlines and retry backoff, and assess persistent-failure isolation before introducing circuit breakers.
- **Diagnostics.** MUST propagate correlation information and classify failures without logging secrets or full sensitive payloads.

## `backend/internal/platform/firestore/key/`

- **Identifiers and serialization.** MUST preserve canonical resource-key and signed-cursor encodings, including existing normalization and delimiter rules.
- **Query security.** MUST bind cursors to their documented resource, actor and channel scope; reject malformed or foreign cursors.
- **Compatibility.** MUST verify old cursor/key behavior and collision boundaries before changing formats; version an incompatible encoding explicitly.

## `backend/internal/platform/grpcstatus/`

- **Error semantics.** MUST translate SDK status codes into existing typed cancellation, deadline and availability outcomes without exposing provider internals.
- **Retry safety.** MUST keep caller cancellation distinct from retryable dependency failures and preserve wrapped-error identity.
- **Testing.** MUST cover changed status mappings directly; do not classify every transport failure as a safe retry.

## `backend/internal/transport/`

- **Boundaries.** MUST translate verified requests/events into use-case calls; transport adapters do not own domain authorization or state transitions.
- **Errors and data.** MUST preserve the owning contract's serialization, identifiers and stable failure semantics across entry points.
- **Diagnostics.** MUST correlate request/event handling using structured metadata while excluding credentials and sensitive bodies.

## `backend/internal/transport/http/`

- **HTTP and errors.** MUST emit the shared contract error envelope, status codes and recovery metadata; never infer success from a partially written response.
- **Security and caching.** MUST apply existing body limits, authentication/CSRF rules and private-response cache policy at the proper boundary.
- **Reliability.** MUST propagate request cancellation/deadlines and verify malformed, rejected and unavailable response paths.

## `backend/internal/transport/http/api/`

- **Contract wiring.** MUST register generated operations with their owning adapters and capability inventory; removed frontend calls require consumer-inventory review.
- **Generation and types.** MUST change OpenAPI sources before regenerating handlers/types; do not patch generated signatures or bypass dispatch assertions.
- **Authorization.** MUST keep middleware identity checks separate from resource permissions rechecked by the use case.

## `backend/internal/transport/http/browser/`

- **Routes and compatibility.** MUST keep static exports, route registration and legacy redirects synchronized; preserve allowed query parameters without permitting external redirects.
- **Security and caching.** MUST preserve CSP, private-page cache rules and asset integrity while changing page dispatch.
- **Acceptance.** MUST build static assets and test real registered routes, redirects and missing-page behavior before removing an old page.

## `backend/internal/transport/http/geolocation/`

- **Validation and privacy.** MUST validate coordinates, precision and timestamps against the location contract without treating client data as proof of business authorization.
- **HTTP semantics.** MUST preserve public diagnostic credential, body-limit and error semantics independently of authenticated attendance operations.
- **Testing.** MUST verify malformed, stale and disallowed requests; do not add location persistence or retention policy through diagnostic changes.

## `backend/internal/transport/line/`

- **Events and identity.** MUST process only verified LINE events through existing schemas and stable event identifiers; observed users do not automatically gain resource access.
- **Reliability and idempotency.** MUST respect reply-token lifetime and duplicate-delivery rules, with bounded outgoing calls and explicit failure handling.
- **Logging.** MUST correlate event handling without recording tokens or unnecessary message content.

## `backend/internal/transport/line/navigation/`

- **UI and identifiers.** MUST keep menu actions, aliases and deep-link targets consistent with shared menu definitions and registered page routes.
- **Events and safety.** MUST route verified postbacks to existing operations; navigation hints never replace session or resource authorization.
- **Compatibility.** MUST test old entry links and menu-action mappings when changing navigation, without publishing remote menus as a side effect.

## `docs/architecture/`

- **Boundaries and dependencies.** MUST describe owned data, dependency direction and integration boundaries using the current module structure and contracts.
- **Consistency and evolution.** MUST make transaction, concurrency, event and migration assumptions explicit; record consequential decisions through the existing ADR process.
- **Evidence.** MUST label current/proposed diagrams and link their implementation or acceptance sources, rather than presenting a target architecture as deployed behavior.

## `docs/development/`

- **Development workflow.** MUST define a reviewable milestone, its applicable checks and Definition of Done before describing a delivery procedure.
- **Single-source configuration.** MUST reference the declared toolchain, command definitions and lockfiles rather than maintaining duplicate version tables.
- **Verification.** MUST separate local tests, contract checks, browser/emulator evidence and deployment acceptance, with reproducible prerequisites.

## `docs/development/tooling/`

- **Tool structure and configuration.** MUST file guidance by the tool's actual responsibility and link version/configuration authority instead of copying it.
- **Reproducibility and security.** MUST use pinned commands with explicit working directories; document required credentials without storing credential values.
- **Evidence.** MUST distinguish installed, configured, connected and behaviorally verified states, including timeout and recovery limitations.

## `docs/development/tooling/agents/`

- **Integration and authorization.** MUST describe the actual host-exposed tools and required resource scopes; tool visibility or a successful handshake grants no extra write authority.
- **Reliability and diagnostics.** MUST record bounded startup/call timeouts and distinguish configuration, protocol, credential and service failures without exposing tokens.
- **Compatibility.** MUST verify version-matched official sources before changing MCP, CLI or sign-in instructions; preserve their distinct credential lifecycles.

## `docs/domains/`

- **Ubiquitous language.** MUST keep entity, resource, actor and operation terms consistent with the owning bounded context and its API contract.
- **Business operations.** MUST state authorization, legal transitions, transaction/version conditions, retry identity and lifecycle effects for the changed operation.
- **Acceptance and evolution.** MUST link representative unit/integration/end-to-end evidence and describe persisted-data or compatibility migrations separately from target behavior.

## `docs/operations/`

- **Deployment and versions.** MUST record the exact artifact/revision, target environment, health evidence and rollback path for an authorized deployment.
- **Migration and cache.** MUST describe expand-contract ordering, index/data prerequisites and any required cache invalidation before retiring old structures.
- **Security and observability.** MUST keep least-privilege identities and secret references explicit; use correlated structured logs, traces and bounded metrics for diagnosis.

## `frontend/miniapp/`

- **Structure and builds.** MUST keep route composition in Pages Router and preserve static export; never make a page depend on unavailable per-request server execution.
- **Components and forms.** MUST reuse protected components through public props and application composition, with accessible field labels and error feedback.
- **Configuration and types.** MUST keep public runtime configuration separate from secrets and use the pinned workspace dependency/type-checking workflow.

## `frontend/miniapp/src/`

- **Structure.** MUST keep routes, identity, platform integration, business UI and shared mechanics in their existing responsibility directories.
- **Types and dependencies.** MUST import generated contracts through the shared API client; avoid cross-feature imports of private state or duplicate transport types.
- **Lifecycle.** MUST cancel obsolete reads and clear identity/resource-scoped data rather than allowing an old response to overwrite a new scope.

## `frontend/miniapp/src/business/`

- **Authorization and state.** MUST render server-provided capabilities and results without reconstructing ownership or domain transition policy in the browser.
- **Forms and recovery.** MUST reuse existing schema-backed fields and shared field/error mappings; preserve the original request and identity after an unknown write, and block navigation where an unresolved operation requires it.
- **Query scope.** MUST keep filters, cursors and projections tied to the selected authorized resource and reject late results from a prior scope.
- **UI acceptance.** MUST verify affected loading, empty, denied and retry states through public controls, including focus and stale-response cleanup.

## `frontend/miniapp/src/business/attendance/`

- **Data, scope and state.** MUST keep launch tokens, personal records, corrections and review authority distinct while displaying server-calculated attendance values.

## `frontend/miniapp/src/business/attendance/approvals/`

- **Data, scope and state.** MUST keep pending review selection, detail and action state synchronized; a capability change must remove stale review controls and private results.

## `frontend/miniapp/src/business/attendance/corrections/`

- **Data, scope and state.** MUST preserve the exact correction body and retry key after an unknown submission; history pagination must not grant review authority.

## `frontend/miniapp/src/business/attendance/leave/`

- **Data, scope and state.** MUST distinguish registration from approval, preserve original uncertain submissions and handle expired launch links through the documented recovery path.

## `frontend/miniapp/src/business/attendance/punch/`

- **Data, scope and state.** MUST respect one active location/punch attempt and its request identifier; abort obsolete watchers without interpreting cancellation as a confirmed punch.

## `frontend/miniapp/src/business/attendance/records/`

- **Data, scope and state.** MUST keep month and cursor navigation scoped to the personal records contract; render backend worktime results instead of recalculating summaries.
- **Queries and time.** MUST keep the selected business month and failed-page cursor together, clear obsolete private results and respect the records API error/cache semantics.

## `frontend/miniapp/src/business/discussion/`

- **Data, scope and state.** MUST keep reply/edit drafts and pagination independent, and use separate server capabilities for editing and classification.

## `frontend/miniapp/src/business/issue/`

- **Data, scope and state.** MUST use server capabilities for content and classification controls; display Issue states without equating closure with approval.

## `frontend/miniapp/src/business/journal/`

- **Data, scope and state.** MUST keep repository, business date and journal link context aligned; render server progress projections without exposing inaccessible source metadata.

## `frontend/miniapp/src/business/object/`

- **Data, scope and state.** MUST keep exact quantity input strings and parent/resource identities intact; display server conversion defaults and progress-reset requirements.

## `frontend/miniapp/src/business/project/`

- **Data, scope and state.** MUST keep Project Owner, destination-copy Owner and source repository selections distinct; hiding a source or view field does not grant source access.

## `frontend/miniapp/src/business/repository/`

- **Data, scope and state.** MUST keep repository selection and navigation guards synchronized; invalidate stale scopes on identity changes and retain unresolved write protection.

## `frontend/miniapp/src/business/repository/insights/`

- **Data, scope and state.** MUST display only scoped server statistics and explicit unavailable/empty states; a missing projection is not a zero-valued business result.
- **Queries and cache.** MUST cancel obsolete repository/actor reads, retry the same scoped query and clear denied results; do not add write or form state to a read-only insight.

## `frontend/miniapp/src/business/repository/labels/`

- **Data, scope and state.** MUST preserve selected label IDs while paging authorized options; retry the failed cursor and clear inaccessible options after denial.

## `frontend/miniapp/src/business/repository/milestones/`

- **Data, scope and state.** MUST use server-returned milestone IDs and capability flags; distinguish an unloaded option from an invalid selection and preserve failed-page state.

## `frontend/miniapp/src/business/subobject/`

- **Data, scope and state.** MUST reuse Object-owned types and editing flows for descendants; keep parent scope, branch focus and navigation locks distinct.

## `frontend/miniapp/src/diagnostics/`

- **Purpose and privacy.** MUST keep diagnostic requests separate from business operations and avoid retaining sensitive samples beyond the documented use.
- **Errors and lifecycle.** MUST present permission denial, timeout and cancellation distinctly, and release watchers or requests on restart/unmount.
- **Verification.** MUST report browser/device limitations separately from unit or contract test results.

## `frontend/miniapp/src/diagnostics/geolocation/`

- **Contract and types.** MUST send location samples through the diagnostic client using contract units and timestamps; never treat a sample as business authorization.
- **Reliability and privacy.** MUST keep one active watcher/request, enforce its timeout and ignore late responses without persisting location history.
- **UI states.** MUST present permission denial, invalid samples, cancellation and successful server validation distinctly.

## `frontend/miniapp/src/identity/`

- **Authentication.** MUST keep login, session renewal and logout in the existing identity flow; backend capability data remains the authority for resource actions.
- **Cache and lifecycle.** MUST invalidate private client data on logout, expiry or actor change; abort old requests and prevent late responses from restoring a prior identity.
- **Errors and security.** MUST preserve stable authentication/recovery metadata through request helpers and never place credentials in URLs or logs.

## `frontend/miniapp/src/navigation/`

- **Routes and identifiers.** MUST use registered route/query conventions and encode resource identifiers without embedding credentials.
- **State and authorization.** MUST navigation visibility may reflect capabilities but cannot authorize a resource; preserve confirmation for dirty or uncertain writes.
- **Compatibility.** MUST test deep links, reload and back navigation when changing routes or selected scope.

## `frontend/miniapp/src/navigation/home/`

- **UI and routes.** MUST keep home actions semantically named, keyboard-accessible and aligned with the actual page entry points.
- **Identity and scope.** MUST preserve intended deep-link targets through login and do not confuse organization selection with resource authorization.
- **Acceptance.** MUST check empty membership, denied scopes and stale account data through the visible navigation controls.

## `frontend/miniapp/src/pages/`

- **Structure and routing.** MUST keep pages as thin compositions of Shell and feature modules; use validated route parameters and existing resource identifiers.
- **Build compatibility.** MUST synchronize static exports, Go route registration and legacy redirects when adding or retiring a page.
- **Identity and lifecycle.** MUST preserve intended entry targets through login and reset resource-scoped UI on route/actor changes without discarding protected writes.

## `frontend/miniapp/src/pages/miniapp/`

- **Structure and routing.** MUST keep pages as thin compositions of Shell and feature modules; use validated route parameters and existing resource identifiers.
- **Build compatibility.** MUST synchronize static exports, Go route registration and legacy redirects when adding or retiring a page.
- **Identity and lifecycle.** MUST preserve intended entry targets through login and reset resource-scoped UI on route/actor changes without discarding protected writes.

## `frontend/miniapp/src/platform/`

- **Dependencies and types.** MUST wrap platform I/O through the existing integration boundary and expose only the types callers actually need.
- **Identity and security.** MUST keep provider profiles, verified sessions and business permissions distinct; do not expose server credentials.
- **Lifecycle.** MUST handle unsupported environments, cancellation and cleanup without duplicating platform initialization in feature pages.

## `frontend/miniapp/src/platform/line/`

- **Identity and integration.** MUST use the existing LIFF initialization/login boundary; provider profile data does not establish backend authorization.
- **Lifecycle and reliability.** MUST keep initialization and callbacks safe across remounts, and preserve intended navigation through supported login flows.
- **Verification.** MUST distinguish SDK mocks from actual LINE WebView/device acceptance and test unsupported or rejected SDK states.

## `frontend/miniapp/src/shared/`

- **Component boundaries.** MUST share proven presentation and interaction mechanics through small typed props; do not introduce business authorization into shared UI.
- **Forms and feedback.** MUST keep labels, field/error associations and loading/empty/error announcements consistent across callers.
- **State ownership.** MUST keep derived display values out of synchronized state and clean up obsolete asynchronous work at the owning lifecycle boundary.

## `frontend/miniapp/src/shared/hooks/`

- **Types and dependencies.** MUST give hooks explicit inputs/results and keep their dependency scope mechanical rather than feature-policy-specific.
- **Lifecycle and state.** MUST clean up listeners, timers and requests; verify remounts and late completions without duplicating derived state.
- **Component compatibility.** MUST preserve protected upstream hooks and public caller behavior; compose application-specific behavior outside protected files.

## `frontend/miniapp/src/shared/http/`

- **Errors and retry identity.** MUST preserve stable API codes, versions and recovery metadata; cancellation or a lost response is not proof that a write failed.
- **Queries and cache.** MUST respect response cache policy, scope reads by actor/resource and discard stale results when that scope changes.
- **Reliability.** MUST bound requests, propagate abort signals and retry writes only according to the original request/idempotency contract.

## `frontend/miniapp/src/shared/location/`

- **Types and time.** MUST preserve contract coordinate units, precision and sample timestamps; client checks provide feedback rather than trusted location authorization.
- **Lifecycle and reliability.** MUST stop watchers/timers on completion, cancellation or unmount and ignore obsolete callbacks.
- **Privacy.** MUST keep samples purpose-limited and do not add retention, background tracking or sensitive diagnostic logging implicitly.

## `frontend/miniapp/src/shared/time/`

- **Time and serialization.** MUST distinguish RFC 3339 UTC instants from business dates and format display values in the contract's explicit time zone.
- **Validation.** MUST reject ambiguous or invalid date/month inputs instead of relying on device-local parsing or silent coercion.
- **Testing.** MUST cover time-zone boundaries and round trips; formatting helpers must not independently decide attendance or payroll rules.

## `frontend/miniapp/src/shared/ui/`

- **Design-system compatibility.** MUST preserve protected component props, variants, semantics and upstream accessibility behavior; this guidance does not authorize primitive changes.
- **Forms and interaction.** MUST compose labels, descriptions, errors, focus and disabled states through public APIs rather than embedding domain rules in primitives.
- **Types and verification.** MUST keep generated component dependencies and types consistent and verify keyboard/responsive behavior for explicitly authorized updates.

## `frontend/miniapp/src/shell/`

- **Structure and UI.** MUST keep application framing, navigation and identity composition separate from feature forms and protected sidebar primitives.
- **Scope and state.** MUST preserve intended routes and dirty/unknown-write guards across sidebar, account and organization changes.
- **Accessibility.** MUST maintain semantic headings, skip navigation, focus behavior and responsive layouts through existing component APIs.

## `frontend/miniapp/src/shell/sidebar-07/`

- **Upstream compatibility.** MUST preserve the protected sidebar block and its public prop/event contracts; compose feature-specific behavior in the application shell.
- **UI and identity.** MUST use real route/account data and accessible navigation names; do not restore demo identities or fake destinations.
- **Verification.** MUST for explicitly authorized changes, compare pinned upstream output and verify keyboard, drawer and focus behavior without blanket overwrites.

## `frontend/miniapp/src/styles/`

- **UI and formatting.** MUST use existing design tokens, theme variables and scoped utilities; do not override protected component internals globally.
- **Compatibility.** MUST check narrow layouts, long content, contrast and focus visibility when changing shared styles.
- **Builds.** MUST preserve the existing stylesheet/toolchain entry points and verify affected static output without adding a parallel styling system.

## `frontend/miniapp/tests/`

- **Test boundaries.** MUST exercise observable behavior through public controls and contract-shaped fixtures rather than reproducing component internals.
- **Isolation and lifecycle.** MUST restore browser mocks, sessions and mounted roots between cases; explicitly verify denial, cancellation and unknown-write retries where relevant.
- **Acceptance evidence.** MUST keep unit/DOM, contract, browser and real-device claims distinct and bind results to the source revision tested.

## `scripts/build/`

- **Reproducible builds.** MUST use locked inputs, explicit build contexts and existing image stages; verify the runtime contract rather than only compilation.
- **Security.** MUST pass required trust material through the established secret mechanism and keep credentials out of layers, logs and build arguments.
- **Artifact lifecycle.** MUST identify the built revision/image and cleanup temporary artifacts without publishing or deploying implicitly.

## `scripts/check/`

- **Automated enforcement.** MUST validate the authoritative configuration/schema relationships and fail nonzero for drift, missing consumers or invalid definitions.
- **Test integrity.** MUST keep checks deterministic and offline where specified; do not relax assertions, exclusions or baselines to hide a regression.
- **Diagnostics.** MUST report the concrete offending path/rule while redacting secrets and leaving source files unchanged.

## `scripts/contract/`

- **API and generation.** MUST verify operation ownership, route dispatch, registered consumers and generated signatures against the same contract revision.
- **Compatibility.** MUST update consumer inventory when a call is added or removed while retaining supported public operations and stable errors.
- **Acceptance.** MUST reproduce invalid contract/dispatch cases and preserve fail-closed checks; generated output is verified rather than patched.

## `scripts/db/`

- **Data migration.** MUST match indexes/migrations to the actual query and target project; use expand-contract ordering for incompatible persisted changes.
- **Consistency and recovery.** MUST state repeatability, partial-failure behavior and recovery prerequisites before changing data or index lifecycle.
- **Safety and verification.** MUST use the established emulator/dry-run path where available and require authorized targets before remote mutation.

## `scripts/dev/`

- **Configuration.** MUST load declared non-secret settings through the existing environment entry point without duplicating defaults across scripts.
- **Lifecycle.** MUST manage child processes, readiness, cancellation and shutdown explicitly; a listening port alone does not prove feature acceptance.
- **Diagnostics.** MUST preserve original command failures and emit usable, redacted startup information.

## `scripts/generate/`

- **Code generation.** MUST derive outputs only from declared schemas/assets/sources using pinned tools, then compare artifact paths and consumer references.
- **Determinism and types.** MUST preserve generated names, serialization and target-language signatures; repeated runs on identical inputs must not introduce unexplained drift.
- **Lifecycle.** MUST keep temporary output isolated and clean it after failure; never hand-edit generated files as a substitute for fixing their source.

## `scripts/mcp/`

- **Integration contracts.** MUST preserve MCP framing, initialization and tool schema/version boundaries; distinguish transport probes from host tool exposure.
- **Reliability and cleanup.** MUST bound process/call lifetimes, close streams and terminate spawned process groups on failure or cancellation.
- **Identity and diagnostics.** MUST check the appropriate credential scope without printing tokens, and keep read-only probes separate from authorized writes.

## `scripts/ops/`

- **Target configuration.** MUST identify project, region, service and revision explicitly before inspecting operational state.
- **Observability.** MUST correlate structured logs, traces, metrics and health evidence for the same revision rather than mixing historical results.
- **Security.** MUST keep inspection read-only and redact secrets; diagnosis does not authorize changing IAM, deployment or resources.

## `scripts/perf/`

- **Measurement.** MUST profile a defined workload/package and compare before/after results under matching conditions; avoid optimization based only on code shape.
- **Observability.** MUST use interpretable latency/allocation/CPU evidence and bounded metric dimensions rather than resource IDs as labels.
- **Artifact lifecycle.** MUST keep profiles in the declared local artifact directory and exclude sensitive payloads or credentials.

## `scripts/release/`

- **Versions and compatibility.** MUST apply the existing compatibility policy and Semantic Versioning where the published artifact uses it; identify the exact tested revision.
- **Delivery workflow.** MUST keep independently verified milestones and purpose-specific Conventional Commit messages; record review and Definition of Done evidence.
- **Deployment and recovery.** MUST verify target health and rollback prerequisites for authorized releases, distinguishing a repository push from production acceptance.

## `scripts/setup/`

- **Pinned configuration.** MUST install declared tool/dependency versions, validate platform/checksums and preserve manifest/lockfile consistency.
- **Idempotency.** MUST make repeated setup safe and failures explicit; do not silently upgrade tools or leave partially configured resources.
- **Security and lifecycle.** MUST keep writable state in the declared workspace paths and credentials out of installation logs or tracked configuration.

## `scripts/test/`

- **Test isolation.** MUST use controlled local MCP/container fixtures with explicit timeouts and process cleanup; do not depend on production credentials.
- **Contract evidence.** MUST assert observable route, protocol and runtime behavior for success, denial and failure without mirroring implementation details.
- **Delivery.** MUST report failed, passed and skipped cases separately and bind acceptance results to the tested revision.

## `scripts/tooling/`

- **Configuration authority.** MUST resolve commands and versions through the declared mise configuration and existing wrapper without hidden fallback upgrades.
- **Reliability and errors.** MUST preserve exit status, cancellation and bounded diagnostics across wrapper/subprocess boundaries.
- **Security and reproducibility.** MUST separate CLI availability from authentication and MCP exposure, recording exact versions without leaking credentials.
