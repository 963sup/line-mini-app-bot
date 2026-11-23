<a id="backendinternalbusinessobject-工作指引"></a>

# backend/internal/business/object agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [object.go](domain/entity/object.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: object and subobject ownership](../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

<a id="變更工項進度的維護與驗收"></a>

## Maintaining and verifying work-item progress changes

Reject work items with total 5 and completed 8. Equivalent decimal representations must produce the same completion state.

When changing this rule, check the [decision flow](../../../../docs/domains/repository/object/README.md#decision-flow) and [representative tests](application/progress_test.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Validate exact quantities, parent scope and legal progress transitions in the domain; record history only for committed, authorized version changes.
