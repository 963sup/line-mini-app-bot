<a id="frontendminiappsrcbusinessattendancerecords-工作指引"></a>

# frontend/miniapp/src/business/attendance/records agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: shared state rules/event reconstruction](../../../../../../docs/domains/attendance/worktime.md#共用狀態規則與事件重建)

## Applicable engineering conventions

- **Data, scope and state:** Keep month and cursor navigation scoped to the personal records contract; render backend worktime results instead of recalculating summaries.
- **Queries and time:** Keep the selected business month and failed-page cursor together, clear obsolete private results and respect the records API error/cache semantics.
