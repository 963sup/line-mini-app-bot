<a id="backendinternalbusinessdiscussion-工作指引"></a>

# backend/internal/business/discussion agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [application service](application/service.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: minimal boundaries and verification](../../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

<a id="調整討論分類的維護與驗收"></a>

## Maintaining and verifying discussion categorization

Members with categorization-only permissions may recategorize, but requests also changing content must reject the unauthorized changes.

When changing this rule, check the [decision flow](../../../../docs/domains/repository/discussion/README.md#decision-flow) and [representative tests](domain/entity/classification_test.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Check content-edit and classification authority separately, and preserve comment/topic versions and lifecycle semantics on retries.
