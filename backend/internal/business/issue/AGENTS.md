<a id="backendinternalbusinessissue-工作指引"></a>

# backend/internal/business/issue agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [issue.go](domain/entity/issue.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="更新議題的維護與驗收"></a>

## Maintaining and verifying Issue updates

Creators may reopen closed Issues. Members with categorization-only permissions cannot change content through categorization.

When changing this rule, check the [decision flow](../../../../docs/domains/repository/issue/design.md) and [representative tests](domain/entity/issue_test.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Apply Issue-owned transition and edit policies without importing attendance approval states or granting content edits through classification.
