<a id="frontendminiappsrcbusinessrepositoryinsights-工作指引"></a>

# frontend/miniapp/src/business/repository/insights agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: space loading sequence](../../../../../../docs/domains/repository/README.md#空間載入時序)

## Applicable engineering conventions

- **Data, scope and state:** Display only scoped server statistics and explicit unavailable/empty states; a missing projection is not a zero-valued business result.
- **Queries and cache:** Cancel obsolete repository/actor reads, retry the same scoped query and clear denied results; do not add write or form state to a read-only insight.
