<a id="frontendminiappsrcbusinessattendancepunch-工作指引"></a>

# frontend/miniapp/src/business/attendance/punch agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: location preflight, state and sequence](../../../../../../docs/domains/attendance/design.md)

## Applicable engineering conventions

- **Data, scope and state:** Respect one active location/punch attempt and its request identifier; abort obsolete watchers without interpreting cancellation as a confirmed punch.
