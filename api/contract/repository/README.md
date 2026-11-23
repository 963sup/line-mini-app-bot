<a id="apicontractrepository-目錄導覽"></a>

# api/contract/repository directory guide

Defines APIs for Repositories, spaces, labels, milestones and statistics.

Maintain this capability's operations and schemas. Root specifications compose them; public formats are generated from contracts.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [labels.yaml](labels.yaml), [milestones.yaml](milestones.yaml), [repositories.yaml](repositories.yaml), [spaces.yaml](spaces.yaml) and [statistics.yaml](statistics.yaml). Read the [local guidance](AGENTS.md) before editing.

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: space loading sequence](../../../docs/domains/repository/README.md#空間載入時序)

Repository lists/details return `groupId` only for group-owned repositories. The MINI App uses it for attendance entry; attendance APIs still enforce their own eligibility rules.

Project creation uses `GET /api/v1/repositories?ownerId=…` to list sources under the same owner. Each page reads the actor's candidates, then filters by owner, active status and live permissions; empty pages may still have successors. Cursors are bound to the original actor and query scope. Omitting `ownerId` retains cross-owner queries.

[Parent directory guide](../README.md)
