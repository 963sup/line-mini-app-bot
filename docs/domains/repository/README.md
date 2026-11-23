# Personal and group repositories

**Implemented** · [Domain design](#領域設計) · [Core flow](#decision-flow) · [Decision implementation](../../../backend/internal/business/repository/domain/entity/repository.go) · [Code](../../../backend/internal/business/repository/README.md) · [Guidance](../../../backend/internal/business/repository/AGENTS.md)

<a id="領域設計"></a>

## Domain design

**Subdomain classification:** core candidate for construction resource management.

**Language:** repositories have IDs, owners and active state; spaces are internal locations, separate from cross-repository projects.

**Context/contracts:** Issue, Discussion and Object use Repository scope/state, checking current repository facts within write transactions.

Issue, Discussion and Object are capabilities within the Repository context. Repository owns ID/version limits, labels, milestones and spaces. Issue/Discussion own content updates and validate referenced catalogs within write transactions. Object owns hierarchy rules while reusing Repository ID/space contracts.

**Acceptance:** matching IDs do not preserve authorization after owner changes or non-ACTIVE state.

[Implementation/tests](../../../backend/internal/business/repository/domain/entity/repository.go) · [Strategic classification criteria](../../references/ddd/README.md)

Specifications for Repositories and spaces, labels, issues, discussions and other capabilities.

This page owns Repository scope/authorization; capability documents own details.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points
[Containers/LINE journeys](journey.md) explains responsibilities, target navigation, wiring and acceptance.

Start with [spaces.md](spaces.md), [Issue specifications](issue/design.md) and [organization/team ownership](../../contracts/organization-team-ownership.md).

<a id="repository-capabilities"></a>

<a id="容器內能力"></a>

## Container capabilities

Capabilities share Repository scope but own operation/authorization rules. Documents are grouped by container; code remains separated by capability.

| Capability | Specification/acceptance |
|---|---|
| Spaces | [Space/ID ownership](spaces.md) |
| Issue | [Entry/core flow](issue/README.md) · [Design/contracts](issue/design.md) |
| Discussion/comments | [Entry/core flow](discussion/README.md) · [Design/contracts](discussion/design.md) |
| Object/subobjects | [Entry/core flow](object/README.md) · [Objects](object/design.md) · [Templates](object/template.md) |
| Labels | [Catalog/references](labels.md) |
| Milestones | [Specification](milestones.md) |
| Shared numbering | [Numbering rules](numbering.md) |
| Insights | [Overview/Pulse](insights.md) |
| Wiki | [Pages/revisions](wiki.md) |
| Container acceptance | [Completion boundaries](acceptance.md) |

`Object.parentId` represents subobjects governed by [object specifications](object/design.md); no separate Subobject domain documents. Cross-repository [Project](../project/README.md) and independent [Journal](../journal/README.md) retain their domains.

<a id="個人與群組-repository"></a>

```mermaid
flowchart LR
  h["Personal and group Repository"]
  h --- n0["Repository stores construction data and spaces.<br/>See Project specifications for cross-repository references and management."]
  n0 --- n0l0["Repository spaces"]
  click n0l0 "spaces.md"
  n0 --- n0l1["Project specification"]
  click n0l1 "../project/project.md"
```


<a id="範圍與資料"></a>

```mermaid
flowchart LR
  h["Scope and data"]
  h --- n0["One bot-managed space per LINE group, with separate id/ownerId.<br/>LINE events bind personal/group owners; no manual creation, rebinding or deletion.<br/>Container management grants no review or other-person attendance access."]
  h --- n1["Verified follow events create personal repositories by UserID and personal candidates."]
  h --- n2["channels/{channelId}/repositories/{id} stores id, ownerId, name, status,<br/>version, createdAt/updatedAt plus internal last-event time/ID.<br/>repositoryOwnerBindings/{ownerId} and random-ID containers are created atomically;<br/>retries/concurrency reuse owner bindings. ACTIVE/INACTIVE tracks bot group entry.<br/>Leaving retains data; rejoining reuses IDs. Old events cannot replace newer state;<br/>inactivation wins equal timestamps. Versions increment only for public state/name changes.<br/>Names are join/first-interaction snapshots; failed queries allow blanks and later retries.<br/>UI falls back to Group space; no live rename synchronization is claimed."]
```

<a id="操作契約"></a>

```mermaid
flowchart LR
  h["Operation contracts"]
  h --- n0["Actor / Resource / Action: verified bot webhook;<br/>group Repository; Observe to create/update from group events."]
  h --- n1["Context / Authority: Messaging API channel, event time/ID;<br/>active-mode join, leave, message, postback and member-join group events only."]
  h --- n2["Policy / Decision: valid groupId; transaction reads unique binding/container,<br/>preventing duplicates/stale overwrites. Do not consume events;<br/>summary and attendance continue normally."]
  h --- n3["Result: atomically create/update containers and user candidates;<br/>do not move or rewrite summary/attendance history."]
  h --- n4["Retries / concurrency: unique-binding transaction, fixed new ID outside retries;<br/>same event never increments version."]
  h --- n5["Event / Audit: last source event ID/time and public update time;<br/>no tokens, message content or personal profiles."]
```

```mermaid
flowchart LR
  h["Operation contracts"]
  h --- n0["Actor / Resource / Action: cookie-session/bearer-authenticated user;<br/>Repository; List / Get."]
  h --- n1["Context / Authority: verified-event candidates. Personal ownerId matches user;<br/>group repositories query live LINE group-member-profile on every read."]
  h --- n2["Policy / Decision: show only ACTIVE verified containers.<br/>Missing, inactive or unauthorized details return 404.<br/>LINE errors other than 200/404 or invalid profiles return 503, never authorization."]
  h --- n3["Result: non-cacheable basics; no summary content,<br/>other-person attendance or management capabilities."]
  h --- n4["Retries / concurrency: reads have no side effects.<br/>Membership is a current external observation, not an atomic LINE/Firestore snapshot."]
  h --- n5["Event / Audit: ordinary reads store no additional personal data/events."]
```

<a id="api-與前端"></a>

```mermaid
flowchart LR
  h["API and frontend"]
  h --- n0["GET /api/v1/repositories?cursor=... reads up to ten personal candidates<br/>then verifies current eligibility. Empty pages may have nextCursor; continue paging."]
  h --- n1["GET /api/v1/repositories/{id} checks personal ownership/live membership.<br/>Malformed, missing or unauthorized IDs reveal no data."]
  h --- n2["Repository API contracts compose through OpenAPI;<br/>generate Go/TS DTOs without manual edits."]
  n2 --- n2l0["Repository"]
  click n2l0 "../../../api/contract/repository/repositories.yaml"
  n2 --- n2l1["OpenAPI entry point"]
  click n2l1 "../../../api/contract/openapi.yaml"
  h --- n3["Sidebar Spaces expands authorized spaces. Selection shows details at<br/>/miniapp/repositories?selected=... without duplicating the list.<br/>Existing ?id=... and repository/liff.state group entries remain compatible."]
  h --- n4["Normal selection uses clean selected parameters, excluding launch/form queries.<br/>Legacy id/repository group entries read specified details only, never general lists<br/>or another group after failures. Parse IDs only on space pages."]
  h --- n5["Cancel stale list/detail requests. Scope visible state by session epoch, user and scope.<br/>Logout/reverification hides old data even when transport ignores cancellation.<br/>Selection rechecks details through API; cached lists never authorize."]
  h --- n6["Mobile selection closes drawers. Location pages link real spaces,<br/>never fabricate personal space data. Without JavaScript, retain links and notices."]
  h --- n7["Attendance links carry existing group filters without new permissions;<br/>summaries remain group operations."]
  h --- n8["Member-join, personal group messages/postbacks create candidates.<br/>Existing noninteracting members must interact first; candidates are not full rosters."]
```

<a id="空間載入時序"></a>

## Space loading sequence

Current implementation

```mermaid
sequenceDiagram
  participant UI as Sidebar or space page
  participant Context as RepositoryContext
  participant API as Repository API
  alt First sidebar expansion or general space page
    UI->>Context: Load space list
    Context->>API: Query accessible spaces
  else Group-specific entry or selected space
    UI->>Context: Load specified details
    Context->>API: Reverify this space
  end
  API-->>Context: Response
  Context-->>UI: Apply only matching user, session epoch and scope
```

```mermaid
flowchart LR
  h["Space loading sequence"]
  h --- n0["business/repository/RepositoryContext.tsx shares lists, pagination and details.<br/>Empty pages with nextCursor retain Load more spaces.<br/>Failures clear old items and offer retry; candidates/inactive spaces cannot open.<br/>See page/identity lifecycle for identity-change cleanup."]
  click n0 "../../architecture/runtime/react-lifecycle.md#頁面與身分生命週期"
```

<a id="正式空間-ui-基礎雛形"></a>

```mermaid
flowchart LR
  h["Production space UI scaffold"]
  h --- n0["Reuse List/Get contracts and identity isolation. Wide cards show authorized basics<br/>and personal attendance. navigation.ts defines page navigation.<br/>Keep Issues, Discussions, labels/milestones in Repository; Projects use independent<br/>sidebar /miniapp/projects. Legacy entries redirect with explicit Space/source identity,<br/>rather than mounting Project business in Repository View.<br/>Overview uses real data; Issues, Discussions, labels, milestones, Projects and Insights<br/>follow their specifications. Public audit remains unavailable without demo data.<br/>Hide contents on loading, failed verification or logout.<br/>Preserve APIs, bot management, membership and numbering. Author categorization owns labels.<br/>Respect scaffold boundaries and authorized slices without expanding management scope.<br/>Overview integration does not prove other capabilities or real LINE device acceptance."]
  n0 --- n0l0["Issues phase one"]
  click n0l0 "issue/design.md"
  n0 --- n0l1["Discussions and comments"]
  click n0l1 "discussion/design.md"
  n0 --- n0l2["Label specification"]
  click n0l2 "labels.md"
  n0 --- n0l3["Milestones"]
  click n0l3 "milestones.md"
  n0 --- n0l4["Project"]
  click n0l4 "../project/project.md"
  n0 --- n0l5["Insights"]
  click n0l5 "insights.md"
  n0 --- n0l6["Basic scaffold"]
  click n0l6 "acceptance.md"
```

<a id="啟用與遷移限制"></a>

```mermaid
flowchart LR
  h["Activation and migration limits"]
  h --- n0["Reuse GOOGLE_CLOUD_PROJECT, FIRESTORE_DATABASE_ID and Messaging API settings;<br/>missing configuration yields 503. MINI App/Messaging API IDs need the same Provider.<br/>Supported trusted events lazily create existing group containers; no production scans/migrations.<br/>Summary/Attendance history retains groupId. Original container slices add containers,<br/>resolution and reads; feature writes follow their specifications without migrating history."]
```

<a id="查證與驗收"></a>

```mermaid
flowchart LR
  h["Verification and acceptance"]
  h --- n0["Official LINE main dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54:<br/>group chats, join/leave/group-member-profile; 404 never authorizes, check profile userId."]
  h --- n1["Context7 /googleapis/google-cloud-go: RunTransaction retries and reads-before-writes,<br/>checked against firestore v1.29.2. Google Developer Knowledge unavailable;<br/>used official Go client docs without cloud deployment/data operations."]
  h --- n2["Context7 /vercel/next.js/v16.4.5 and installed static-export guidance:<br/>retain Pages Router, client fetching and static export."]
  h --- n3["Graphify located webhook/server/identity/group queries.<br/>Scoped Repomix /tmp/repository-before.xml; read omitted guidance directly."]
  h --- n4["Delivery messages list local results. Live LINE events, production Firestore concurrency<br/>and iOS/Android login require separate acceptance; builds do not replace them."]
  h --- n5["Historical local verification passed: static build/types, 12 frontend tests,<br/>OpenAPI generation/compatibility, configuration, Go vet/race/build/lint/format/vulnerabilities.<br/>Temporary controlled-store/identity/LINE smoke covered unauthenticated/nonmember/inactive,<br/>invalid cursors, profile mismatches and 404/429 rejection semantics.<br/>No live Firestore/LINE acceptance or new permanent tests/tools."]
  h --- n6["LINE menu/help offers Spaces through existing LIFF URL with repositories=1.<br/>Home parses normal query/liff.state and routes to spaces."]
  h --- n7["Trusted groupId resolves Repository ID; links use repositories=1&amp;repository=&lt;id&gt;.<br/>Query/liff.state retain targets. Dedicated entries read only details, including reload;<br/>failures do not fall back to lists. Direct chat lists personal candidates.<br/>Verify personal owner/live group membership; changing IDs grants no access."]
  h --- n8["Historical dedicated-entry smoke used real Go routes and controlled LINE/API responses<br/>for normal queries, liff.state and 404. All requested specified details only,<br/>reload stayed scoped and 390px had no horizontal overflow. Not LINE-device acceptance."]
  h --- n9["Construction item tracking follows Objects, separate from Issues.<br/>No automatic PDF imports or third-party edit grants."]
  click n9 "object/design.md"
```

<a id="後續模型設計"></a>

```mermaid
flowchart LR
  h["Future model design"]
  h --- n0["Shared Repository numbering is used by Issue/Discussion creation;<br/>comments consume none. Bot, public queries and membership rules remain."]
  click n0 "numbering.md"
  h --- n1["See business model design for entity/query/collection/Connection naming<br/>and Issue/Discussion ownership proposals. Future references do not alter<br/>existing REST, membership verification or bot-managed behavior."]
  click n1 "../../proposals/platform/domain-model.md"
```

<a id="decision-flow"></a>

<a id="寫入庫內受限資源的判定流程"></a>

## Restricted repository write decision flow

```mermaid
flowchart TD
    start["Write restricted repository resource"] --> gate{"Current owner and active state valid?"}
    gate -->|Yes| ok["Continue resource operation"]
    gate -->|No| stop["Reject stale authorization"]
```

[Entry adapter](../../../backend/internal/business/repository/adapter/http/routes.go) · [Decision rules](../../../backend/internal/business/repository/domain/entity/repository.go) · [Application output port](../../../backend/internal/business/repository/application/port/repository.go) · [Related tests](../../../backend/internal/business/repository/adapter/firestore/labels_test.go)

[applicable guidance](../AGENTS.md) · [parent directory guide](../README.md)
