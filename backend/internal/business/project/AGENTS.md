<a id="backendinternalbusinessproject-工作指引"></a>

# backend/internal/business/project agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [inspections.go](domain/entity/inspections.go), [project.go](domain/entity/project.go), [references.go](domain/entity/references.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="工作方式"></a>

## Working approach

[Current implementation diagram: minimal source reference structure](../../../../docs/domains/project/project.md#最小來源引用骨架)

Project and Repository verify owners separately. Project Owners are users or organizations: verify the logged-in user or organization access respectively. Projects may link multiple Repositories; authorize each source reference separately. Spaces belong to Repositories.

<a id="加入跨庫來源引用的維護與驗收"></a>

## Maintaining and verifying cross-repository references

Check selected sources and project links when adding cross-repository references. Project permissions do not grant source-read access in other repositories.

When changing this rule, check the [decision flow](../../../../docs/domains/project/design.md) and [representative tests](application/reference_source_test.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Authorize the Project Owner and each referenced source independently; copying creates a new resource identity without inheriting unauthorized source access.
