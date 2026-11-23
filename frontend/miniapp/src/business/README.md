<a id="frontendminiappsrcbusiness-目錄導覽"></a>

# frontend/miniapp/src/business directory guide

Use **feature-based architecture**. Colocate screens, forms, state and typed API calls; `pages/` composes routes. Frontends present inputs/results; backends decide business rules/permissions.

Issue, Discussion, comment, Object, Project and Milestone editing uses API `canEdit`. Subobject forms use parent capabilities. Issue/Discussion classification has separate permissions. Journal relationship options come from group queries; browsers select, paginate/cancel stale requests.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [attendance](attendance), [discussion](discussion), [enterprise](enterprise), [issue](issue), [journal](journal), [object](object); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

The [MINI App home route](../pages/miniapp/index.tsx) composes [Organization](organization/OrganizationPage.tsx), [team lists](team/OrganizationTeams.tsx) and [Enterprise provisioning/management](enterprise/Enterprise.tsx). Organization/Team receive the same organization ID but query separate APIs. Enterprise eligibility comes from sidebar results. Organization teams/Enterprise share [Team forms](team/TeamActions.tsx) for management assignment and enterprise assignment/removal, owning request/UI state while backends authorize.

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[Parent directory guide](../README.md)

<a id="下層文件入口"></a>

## Child documentation entry points

- [project](project/README.md)
- [repository](repository/README.md)
- [subobject](subobject/README.md)
