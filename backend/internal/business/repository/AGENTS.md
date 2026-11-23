<a id="backendinternalbusinessrepository-工作指引"></a>

# backend/internal/business/repository agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [numbering.go](domain/entity/numbering.go), [repository.go](domain/entity/repository.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="工作方式"></a>

## Working approach

[Current implementation diagram: space loading sequence](../../../../docs/domains/repository/design.md)

Repository ownership uses LINE UserID (U) or GroupID (C); verify the logged-in user or current group membership accordingly. Project has its own user/organization Owner and optional source links; follow the Project specification rather than applying Repository Owner rules to Projects. Authorize each Project source reference separately. Spaces belong to Repositories.

<a id="寫入庫內受限資源的維護與驗收"></a>

## Maintaining and verifying restricted repository writes

Even with a matching repository ID, changed ownership or non-ACTIVE status invalidates old authorization for restricted operations.

When changing this rule, check the [decision flow](../../../../docs/domains/repository/design.md) and [implementation](domain/entity/repository.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Revalidate repository ownership and active status at restricted writes; list filters and cached capability flags are not write authorization.
