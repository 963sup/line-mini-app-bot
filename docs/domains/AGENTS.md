<a id="領域規則及驗收工作指引"></a>

# Domain rules and acceptance guidance

Inherit [parent guidance](../AGENTS.md); select topics through the [local index](./).

<a id="職責與邊界"></a>

## Responsibilities and boundaries

Cross-module flows/rules for attendance, worktime, weather, summaries and other features.
Describe frontend/API/backend rules together; link implementation modules rather than duplicating specifications. Distinguish implemented, pending acceptance and future work.

<a id="維護與驗證"></a>

## Maintenance and verification

Update indexes/links per [documentation maintenance](../AGENTS.md#驗證).

Use the [operation contract template](../contracts/business-operations.md#功能操作契約範本) when editing feature documents. Separate existing evidence, target behavior and unresolved rules; do not independently change product scope.

Answer six questions: who acts, on what, doing what, with which permissions, under which rules, and with what changes/results? Add Context, Authority, Decision and risk-appropriate Event/Audit details.
Separate system allow/deny from business approve/reject. Use Merge only for adopting proposals.

Follow [naming standards](../standards/naming.md#業務實體集合與-api-命名) for models/contracts.

<a id="文件歸屬"></a>

## Documentation ownership

Each domain keeps its responsibilities, flows, contracts and acceptance criteria in `<domain>/design.md`. Keep distinct topics in named files and link proposals to their source. [Cross-boundary contracts](../contracts/) own shared guarantees. Update inbound links, diagrams and API capability paths when moving specifications.

[Repository design](repository/design.md#repository-capabilities) indexes Issue, Discussion and Object, including subobjects and templates. Project and Journal retain their own design files. [LINE group compatibility](../integrations/line/group-compatibility.md) documents group discovery and membership checks.

## Applicable engineering conventions

- **Ubiquitous language:** Keep entity, resource, actor and operation terms consistent with the owning bounded context and its API contract.
- **Business operations:** State authorization, legal transitions, transaction/version conditions, retry identity and lifecycle effects for the changed operation.
- **Acceptance and evolution:** Link representative unit/integration/end-to-end evidence and describe persisted-data or compatibility migrations separately from target behavior.
