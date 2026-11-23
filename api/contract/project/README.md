<a id="apicontractproject-目錄導覽"></a>

# api/contract/project directory guide

Defines APIs for construction projects, fields, items, source references and inspection submissions.

Maintain this capability's operations and schemas. Root specifications compose them; public formats are generated from contracts.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [fields.yaml](fields.yaml), [inspection-submissions.yaml](inspection-submissions.yaml), [items.yaml](items.yaml), [owner-projects.yaml](owner-projects.yaml), [projects.yaml](projects.yaml) and [references.yaml](references.yaml). Read the [local guidance](AGENTS.md) before editing.

<a id="owner空白專案與複製-api"></a>

## Owner, blank project and copy APIs

Query projects by Owner, which accepts only users or organizations; the backend verifies entity permissions. Copy operations identify the source project, destination Owner and new name, verify source read and destination create permissions, and return the new project ID and destination Owner. Retrying the same request with the same idempotency key returns the same copy; changed parameters with the same key conflict. Copy content follows the [domain specification](../../../docs/domains/project/project.md#複製到另一個-owner); copies can open without an initial source.

Creation may omit defaultRepositoryId to create a project with zero sources. Details then omit that field and return an empty linked-source list. Initial sources still require authorization; existing projects retain their sources. Later repository links use per-repository authorization.

The [create request](owner-projects.yaml) uses copyFrom to specify source Owner and project. Acceptance includes [project entry and cross-repository cases](../../../docs/domains/project/journey.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: minimal source reference structure](../../../docs/domains/project/project.md#最小來源引用骨架)

[Parent directory guide](../README.md)
