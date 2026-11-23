<a id="契約能力索引"></a>

# Contract capability index

Use **modular OpenAPI contracts organized by domain**. Domains own their operations and schemas; public entry point `openapi.yaml` and internal entry point `internal-openapi.yaml` compose them through `$ref`. Assign contracts by operation responsibility, then generate Go and TypeScript through the existing workflow. **Split contracts as needed.** Small domains may keep related operations and schemas in one file; split when capability boundaries or independent maintenance justify it.

Maintain this capability's operations and schemas. Root specifications compose them; public formats are generated from contracts.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with the [capability and implementation inventory](capabilities.json), [internal-openapi.yaml](internal-openapi.yaml), [openapi.yaml](openapi.yaml) and [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Project specification](../../docs/domains/project/project.md) · [Labels](../../docs/domains/repository/labels.md)

Contracts are organized by capability ownership. Public HTTP, events and internal jobs have different entry forms; classification must not create public APIs. Feature specifications still define authorization and business rules.

| Capability or responsibility | Source                                                                                                | Scope                                                                                                                                           |
| ---------------------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Public HTTP composition      | [openapi.yaml](openapi.yaml)                                                                          | API metadata and `$ref` indexes for existing paths and components                                                                               |
| Shared transport contracts   | [common/transport.yaml](common/transport.yaml)                                                        | APIError, Bearer and Cookie security schemes; no business models                                                                                |
| PDF tools                    | [operations/poppler.yaml](operations/poppler.yaml)                                                    | Authenticated PDF text extraction testing without persistence                                                                                   |
| Identity                     | [identity/identity.yaml](identity/identity.yaml)                                                      | MINI App public configuration, current user, session creation/renewal/revocation and DTOs                                                       |
| Repository                   | [repository/repositories.yaml](repository/repositories.yaml)                                          | Container lists, details, pagination and DTOs; no manual creation or management API                                                             |
| Enterprise                   | [enterprises.yaml](enterprise/enterprises.yaml), [types.yaml](enterprise/types.yaml)                  | Enterprise provisioning, details, owners and organization links                                                                                 |
| Organization                 | [context.yaml](organization/context.yaml)                                                             | Organization selection, creation and owner transfer                                                                                             |
| Team                         | [teams.yaml](team/teams.yaml)                                                                         | Team management, organization team queries and enterprise team assignment                                                                       |
| Group                        | [group/groups.yaml](group/groups.yaml)                                                                | Engineering group lists, details and current membership                                                                                         |
| Repository / Space           | [repository/spaces.yaml](repository/spaces.yaml)                                                      | Workspaces within a repository, using its authorization                                                                                         |
| Project                      | [project/owner-projects.yaml](project/owner-projects.yaml)                                            | Personally or engineering-group-owned projects and items with read-only sources; [Project specification](../../docs/domains/project/project.md) |
| Project legacy               | [project/projects.yaml](project/projects.yaml), [items.yaml](project/items.yaml)                      | Compatibility adapters for repository-scoped routes and DTOs; same use cases, no new ownership                                                  |
| Repository / Labels          | [repository/labels.yaml](repository/labels.yaml)                                                      | List/create label catalog; see [labels](../../docs/domains/repository/labels.md)                                                                |
| Issue                        | [issue/issues.yaml](issue/issues.yaml)                                                                | Space issue lists, details, creation and creator updates; see [issues](../../docs/domains/repository/issue/design.md)                           |
| Discussion                   | [discussion/discussions.yaml](discussion/discussions.yaml), [comments.yaml](discussion/comments.yaml) | Topic reads/writes and single-level comments with author editing; see [discussions](../../docs/domains/repository/discussion/design.md)         |
| Attendance / Punch           | [attendance/punch.yaml](attendance/punch.yaml)                                                        | Launch, punch and preflight                                                                                                                     |
| Attendance / Records         | [attendance/records.yaml](attendance/records.yaml)                                                    | Groups, attendance records, monthly summaries and pagination                                                                                    |
| Attendance / Worktime        | [attendance/worktime.yaml](attendance/worktime.yaml)                                                  | Worktime calculation and interval types; no independent HTTP operations                                                                         |
| Attendance / Corrections     | [attendance/corrections.yaml](attendance/corrections.yaml)                                            | Submit/list corrections, status and self-withdrawal; request snapshots, results and shared correction errors                                    |
| Attendance / Approvals       | [attendance/approvals.yaml](attendance/approvals.yaml)                                                | Review details, approval, rejection and review groups; references Corrections request data                                                      |
| Summary                      | [summary/README.md](summary/README.md)                                                                | MessageCollection, Summarization and internal worker; outside public OpenAPI                                                                    |
| Weather                      | [weather/README.md](weather/README.md)                                                                | Existing Forecast LINE event entry point; outside public OpenAPI                                                                                |

<a id="巢狀指引"></a>

## Nested guidance

This directory follows [contract guidance](AGENTS.md) and inherits [API guidance](../AGENTS.md). Capability-specific additions:

| Scope      | Guidance                                           | Status                                                           |
| ---------- | -------------------------------------------------- | ---------------------------------------------------------------- |
| Common     | [AGENTS.md](common/AGENTS.md)                      | Shared transport contracts                                       |
| PDF tools  | [operations/poppler.yaml](operations/poppler.yaml) | Authenticated PDF text extraction testing without persistence    |
| Identity   | [AGENTS.md](identity/AGENTS.md)                    | Existing public contracts                                        |
| Repository | [AGENTS.md](repository/AGENTS.md)                  | Bot-managed group spaces                                         |
| Issue      | [AGENTS.md](issue/AGENTS.md)                       | Basic Issue API; types/comments remain scaffolds                 |
| Discussion | [AGENTS.md](discussion/AGENTS.md)                  | Topics/single-level comments wired; categories remain a scaffold |
| Project    | [AGENTS.md](project/AGENTS.md)                     | Project/items wired; fields/views are scaffolds                  |
| Attendance | [AGENTS.md](attendance/AGENTS.md)                  | Existing attendance contracts                                    |
| Summary    | [AGENTS.md](summary/AGENTS.md)                     | Event and internal job contracts                                 |
| Weather    | [AGENTS.md](weather/AGENTS.md)                     | Existing event entry contracts                                   |

<a id="歸屬與依賴"></a>

## Ownership and dependencies

Follow [business entity, collection and API naming](../../docs/standards/naming.md#業務實體集合與-api-命名): singular entity types, plural resource collections and REST collection paths. Apply documented singleton and public compatibility exceptions. Do not mechanically determine capability names by English plurality.

- Root entry points compose sources without embedding capability paths, schemas or security schemes. Preserve existing paths and component names in their indexes for public compatibility.
- Capability sources must not reference `openapi.yaml`. Common owns shared errors and authentication. Root securitySchemes resolve security requirement names; capability fragments are not standalone OpenAPI documents.
- Capabilities own operations, inputs and results. Reference genuinely shared business data through its owner rather than moving everything into common.
- Attendance dependencies: Approvals → Corrections → Records → Worktime. Approvals also references Records directly; Punch/Leave reference common. Corrections does not depend on Approvals.
- Preserve the name `CorrectionReview` for compatibility, but Corrections owns its content as the request snapshot shared by lists and reviews. The request lifecycle owner also owns `CorrectionDecisionResult`, used for withdrawal and review.
- One capability file may contain related paths and schemas. Do not require one file per endpoint or create empty HTTP contracts for capabilities without public entry points.

<a id="修改與驗證"></a>

## Changes and verification

<a id="最小契約雛形"></a>

### Minimal contract scaffolds

Repository milestones/statistics and Project/items are wired according to [milestones](../../docs/domains/repository/milestones.md), [Project](../../docs/domains/project/project.md) and [Insights](../../docs/domains/repository/insights.md). The files below are marked `x-status: scaffold`, contain only empty `paths` and `components.schemas`, define no endpoints or fields, and are not composed into public/internal entry points. Recording ownership does not establish implementation.

- Issue：[types.yaml](issue/types.yaml)（IssueType）、[comments.yaml](issue/comments.yaml)。
- Discussion：[categories.yaml](discussion/categories.yaml)。

Project fields and views are exposed through project read/update endpoints; see [fields.yaml](project/fields.yaml) and [views.yaml](project/views.yaml) for types.

These are capability fragments rather than independently runnable or callable OpenAPI documents. Specify operations first, then fill paths/schemas, compose them, update the inventory and generated code, and verify compatibility. Existing repositories are still created by the bot.

Change source contracts first, then generate shared Go/TypeScript artifacts through the [API workflow](../README.md). Redocly bundles from root entry points; do not create per-capability clients or generators. Do not refresh compatibility baselines during moves.

For ownership-only reorganizations, first compare the complete JSON structure of `redocly bundle --dereferenced` before/after, then run `pnpm api:check`. Preserve paths, operationId, type names, authorization, errors, nullable and deprecated semantics. Review generated changes; successful generation alone does not prove compatibility.

OpenAPI modularization does not establish complete runtime validation or atomic database transactions. Check HTTP handlers, business rules and transaction boundaries separately.

<a id="工作日誌"></a>

## Work journals

The [Journal contract](journal/journals.yaml) defines publishing, listing and single-entry reads. Feed and Dashboard use the same resource; see [journal design](../../docs/domains/journal/journal.md) for query behavior and production activation requirements.

[Current implementation diagram: basic entry points](../../docs/contracts/http.md#基礎入口)

<a id="下層文件入口"></a>

## Child documentation entry points

- [group](group/README.md)
- [journal](journal/README.md)
- [object](object/README.md)
