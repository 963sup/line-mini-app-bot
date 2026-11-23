<a id="issue-契約指引"></a>

# Issue contract guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [comments.yaml](comments.yaml), [issues.yaml](issues.yaml), [types.yaml](types.yaml), the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Issue operation specification](../../../docs/domains/repository/issue/design.md) · [Business model design](../../../docs/proposals/platform/domain-model.md)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Status: basic Issue APIs are wired. Read the [Issue specification](../../../docs/domains/repository/issue/design.md) first. Types/comments remain scaffolds; [business model design](../../../docs/proposals/platform/domain-model.md) does not replace confirmed specifications.
- Use Issue for entities and issues for collections. Each Issue belongs to an existing Repository. Confirm routes, fields, identifiers and persistence locations in feature specifications.
- Specify Issue actors, permissions, transitions, versions, retries and necessary audits for creation, queries, updates and lifecycle before defining contracts.
- Closing an Issue does not automatically mean construction acceptance or business approval. Do not reuse attendance correction states or review permissions.
- This capability owns IssueComment and Issue relationships; Label, Milestone and other references belong to their owners. Do not duplicate definitions across capabilities or create empty endpoints for unconfirmed operations.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Contract semantics:** Keep Issue transitions and versions separate from comment, label and milestone references; define concurrent edit and retry outcomes.
