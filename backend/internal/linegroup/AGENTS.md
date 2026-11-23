<a id="line-群組相容整合工作指引"></a>

# LINE group compatibility integration guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [group.go](application/query/group.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="查閱受限群組資料的維護與驗收"></a>

## Maintaining and verifying restricted group reads

Users observed in events still need verified group membership to access restricted data, even when listed as candidates.

When changing this rule, check the [decision flow](../../../docs/integrations/line/group-compatibility.md#decision-flow) and [representative tests](application/query/group_test.go).

## Applicable engineering conventions

- **Authorization:** Resolve restricted group access through verified membership; cached candidates and webhook observation do not prove current membership.
- **Caching and queries:** Scope membership/cache keys by the relevant identity and group, with explicit expiry and failure behavior.
- **Compatibility:** Preserve legacy group identifiers while testing that unavailable or changed membership cannot reuse stale authorization.
