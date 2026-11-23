<a id="frontendminiappsrcbusinessattendance-工作指引"></a>

# frontend/miniapp/src/business/attendance agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [launchToken.ts](launchToken.ts), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: location preflight, state and sequence](../../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序)

## Applicable engineering conventions

- **Data, scope and state:** Keep launch tokens, personal records, corrections and review authority distinct while displaying server-calculated attendance values.
