<a id="業務功能"></a>

# Business features

<a id="code-by-domain"></a>
<a id="依領域查程式目錄"></a>
<a id="domain-matrix"></a>

<a id="領域與容器的追蹤矩陣"></a>

## Domain and container traceability matrix

Organize top-level entries by independent domains/containers. The [Repository capability index](repository/README.md#repository-capabilities) routes Issues, Discussions and Objects/subobjects. Grouping does not alter authorization, state or code responsibility.

Follow hexagonal architecture from entry adapters to use cases/implementations, then output ports, storage and decisions. Links show representative paths; implementation links/proposal acceptance do not prove executed tests.

| Domain entry | Status | Core flow/contracts | Implementation, tests or acceptance | Module entry |
|---|---|---|---|---|
| [agent](agent/README.md) | Proposal | [Core flow](../proposals/agent.md#design-diagram) | [Acceptance criteria](../proposals/domain-collaboration.md) | [Backend](../../backend/internal/business/agent/README.md) |
| [attendance](attendance/README.md) | Implemented | [Core flow](attendance/README.md#decision-flow) | [Tests](../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go) | [Backend](../../backend/internal/business/attendance/README.md) · [Frontend](../../frontend/miniapp/src/business/attendance/README.md) · [Public contracts](../../api/contract/attendance/README.md) |
| [enterprise](enterprise/README.md) | Implemented | [Core flow](enterprise/README.md#decision-flow) | [Implementation](../../backend/internal/business/enterprise/adapter/firestore/store.go) | [Backend](../../backend/internal/business/enterprise/README.md) |
| [LINE group compatibility](group/README.md) | Implemented | [Core flow](group/README.md#decision-flow) | [Tests](../../backend/internal/linegroup/application/query/group_test.go) | [Public contracts](../../api/contract/group/README.md) |
| [journal](journal/README.md) | Implemented | [Core flow](journal/README.md#decision-flow) | [Tests](../../backend/internal/business/journal/domain/entity/entry_test.go) | [Backend](../../backend/internal/business/journal/README.md) · [Frontend](../../frontend/miniapp/src/business/journal/README.md) · [Public contracts](../../api/contract/journal/README.md) |
| [knowledge](knowledge/README.md) | Proposal | [Core flow](../proposals/knowledge.md#design-diagram) | [Acceptance criteria](../proposals/domain-collaboration.md) | [Backend](../../backend/internal/business/knowledge/README.md) |
| [Business operations](operations/README.md) | Generic model proposed; PDF extraction implemented | [Core flow](../proposals/operations.md#design-diagram) | [Acceptance criteria](../proposals/domain-collaboration.md) | [Backend](../../backend/internal/business/operations/README.md) · [Frontend](../../frontend/miniapp/src/business/operations/README.md) · [Public contracts](../../api/contract/operations/README.md) |
| [membership](../proposals/membership.md) | 部分實作；獨立模式切換與完整入口仍待交付 | [成員、邀請與資格責任](../proposals/membership.md#領域邊界) | [邀請使用案例](../../backend/internal/business/membership/application/invitations.go) · [Team 資格](../../backend/internal/business/membership/domain/team_membership.go) · [交易接線](../../backend/internal/bootstrap/server.go) | [正常入口與待定規則](../proposals/membership.md#閉環判定與待定規則) |
| [organization](organization/README.md) | Implemented | [Core flow](organization/README.md#decision-flow) | [Implementation](../../backend/internal/business/organization/adapter/firestore/store.go) | [Backend](../../backend/internal/business/organization/README.md) |
| [project](project/README.md) | Implemented | [Core flow](project/README.md#decision-flow) | [Tests](../../backend/internal/business/project/application/reference_source_test.go) | [Backend](../../backend/internal/business/project/README.md) · [Frontend](../../frontend/miniapp/src/business/project/README.md) · [Public contracts](../../api/contract/project/README.md) |
| [repository](repository/README.md) | Implemented | [Core flow](repository/README.md#decision-flow) | [Implementation](../../backend/internal/business/repository/domain/entity/repository.go) | [Backend](../../backend/internal/business/repository/README.md) · [Frontend](../../frontend/miniapp/src/business/repository/README.md) · [Public contracts](../../api/contract/repository/README.md) |
| [summary](summary/README.md) | Implemented | [Core flow](summary/README.md#decision-flow) | [Implementation](../../backend/internal/business/summary/application/summarization/validation.go) | [Backend](../../backend/internal/business/summary/README.md) · [Public contracts](../../api/contract/summary/README.md) |
| [team](team/README.md) | Implemented | [Core flow](team/README.md#decision-flow) | [Implementation](../../backend/internal/business/team/adapter/firestore/store.go) | [Backend](../../backend/internal/business/team/README.md) |
| [weather](weather/README.md) | Implemented | [Core flow](weather/README.md#decision-flow) | [Tests](../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go) | [Backend](../../backend/internal/business/weather/README.md) · [Public contracts](../../api/contract/weather/README.md) |
| [workflow](workflow/README.md) | Proposal | [Core flow](../proposals/workflow.md#design-diagram) | [Acceptance criteria](../proposals/domain-collaboration.md) | [Backend](../../backend/internal/business/workflow/README.md) |

<a id="業務功能-1"></a>

## Business features

Use status above to distinguish implementations/proposals. Domains own language, rules and acceptance; [API contracts](../../api/README.md) own public formats. Validate core/supporting classifications against adoption reasons. For cross-module questions, see [shared adapters](../architecture/runtime/shared-adapters.md), [DDD methods](../references/ddd/README.md) or [Organization/Team target contracts](../contracts/organization-team-ownership.md).

<a id="出勤"></a>

## Attendance

[Basic attendance](attendance/README.md)

[Group boards/personal missing clock-out checks](attendance/board.md)

[Group leave registration](attendance/leave.md)

[Attendance queries/corrections](attendance/queries.md)

[Review screen flow](attendance/queries.md#審核畫面狀態同步)

[Worktime calculation](attendance/worktime.md)

[Correction state diagram](attendance/worktime.md#補卡與人工審核)

<a id="日誌"></a>

## Journal

[Independent journals: Feed/Dashboard scaffold](journal/journal.md)

<a id="專案"></a>

## Project

[Independent personally/group-owned Project](project/project.md)

<a id="儲存庫"></a>

## Repository

[Repository container/capability index](repository/README.md#repository-capabilities)

Repository capabilities include spaces, issues, discussions/comments, objects/subobjects, templates, labels, milestones, shared numbering, insights and Wiki. The Repository entry routes detailed specifications/acceptance to their single sources.

Cross-repository [Project](project/README.md) and independent [Journal](journal/README.md) retain their own entries.

<a id="群聊摘要"></a>

## Group chat summaries

[Basic group summaries](summary/summary.md)

<a id="天氣"></a>

## Weather

[Group location weather](weather/weather.md)

[Domain maintenance rules](AGENTS.md) · [Parent directory guide](../README.md)
