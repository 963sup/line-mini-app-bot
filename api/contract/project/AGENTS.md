<a id="project-契約指引"></a>

# Project contract guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [fields.yaml](fields.yaml), [inspection-submissions.yaml](inspection-submissions.yaml), [items.yaml](items.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Projects and source references](../../../docs/domains/project/project.md) · [Business model design](../../../docs/proposals/platform/domain-model.md)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Status: Project/items are wired per [projects and source references](../../../docs/domains/project/project.md); Project settings read/write fields/views. Use [business model design](../../../docs/proposals/platform/domain-model.md) as a future reference for the full management model.
- Use Project, ProjectItem, ProjectField and ProjectView. GitHub ProjectV2 is a reference only; do not add V2 suffixes on that basis.
- Users or organizations own Projects; LINE groups are not Project Owners in the canonical API. Creation may omit the initial source repository; a Project may link zero or more authorized repositories. References support objects, subobjects, issues and discussions; existing items preserve Issue-reference compatibility.
- Separate item IDs from source Issue IDs, field definitions/options from item field values, and source Issue state from Project Status.
- Table, Board and Roadmap share item data. Views store presentation and query settings; filtering or hiding fields is not authorization.
- Do not automatically add cross-group access, drafts, scheduling, automation, analytics or templates for scaffold needs. Verify source and project permissions separately.

[Current implementation diagram: minimal source reference structure](../../../docs/domains/project/project.md#最小來源引用骨架)

## Applicable engineering conventions

- **Contract semantics:** Use the owning Project contracts for Owner types and source requirements; distinguish item IDs, source IDs and field-option IDs.
