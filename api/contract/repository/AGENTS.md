<a id="bot-管理的群組空間契約"></a>

# Bot-managed group space contracts

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [labels.yaml](labels.yaml), [milestones.yaml](milestones.yaml), [repositories.yaml](repositories.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Group Repository specification](../../../docs/domains/repository/README.md) · [Label specification](../../../docs/domains/repository/labels.md)

Repository ownership uses LINE UserID (U) or GroupID (C); verify the logged-in user or current group membership accordingly. Project has its own user/organization Owner and optional source links; follow the Project specification rather than applying Repository Owner rules to Projects. Authorize each Project source reference separately. Spaces belong to Repositories.

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Read the [group Repository specification](../../../docs/domains/repository/README.md) first. Each group has one space, created and updated by signature-verified bot events.
- Preserve list, detail and pagination contracts. Do not add manual creation, deletion, rebinding, owner or role management.
- Separate Repository ID and ownerId. Interaction candidates are neither a complete membership roster nor authorization. Container access does not automatically grant child business write or review permissions.
- Label catalogs support list/create per the [label specification](../../../docs/domains/repository/labels.md), excluding rename/delete. References use Issue/Discussion author updates. Milestones/statistics follow [milestones](../../../docs/domains/repository/milestones.md) and [repository insights](../../../docs/domains/repository/insights.md). Define operations and ownership before future Member capabilities; interaction candidates are not a members API.
- Child features reference existing Repository ownership. Issue/Discussion writes must not implicitly create containers.

[Current implementation diagram: space loading sequence](../../../docs/domains/repository/README.md#空間載入時序)

## Applicable engineering conventions

- **Contract semantics:** Scope repository and child-resource queries explicitly; returning a repository or capability flag does not authorize a later write.
