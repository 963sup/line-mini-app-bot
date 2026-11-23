<a id="frontendminiappsrcbusinessattendanceleave-工作指引"></a>

# frontend/miniapp/src/business/attendance/leave agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: minimal flow](../../../../../../docs/domains/attendance/leave.md#最小流程)

## Applicable engineering conventions

- **Data, scope and state:** Distinguish registration from approval, preserve original uncertain submissions and handle expired launch links through the documented recovery path.
