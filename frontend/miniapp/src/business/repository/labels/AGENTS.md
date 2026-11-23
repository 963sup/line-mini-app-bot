<a id="frontendminiappsrcbusinessrepositorylabels-工作指引"></a>

# frontend/miniapp/src/business/repository/labels agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Picker.tsx](Picker.tsx), [View.tsx](View.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: space loading sequence](../../../../../../docs/domains/repository/README.md#空間載入時序)

## Applicable engineering conventions

- **Data, scope and state:** Preserve selected label IDs while paging authorized options; retry the failed cursor and clear inaccessible options after denial.
