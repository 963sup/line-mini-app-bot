<a id="frontendminiappsrcbusinessattendanceapprovals-工作指引"></a>

# frontend/miniapp/src/business/attendance/approvals agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: review screen state synchronization](../../../../../../docs/domains/attendance/queries.md#審核畫面狀態同步)

## Applicable engineering conventions

- **Data, scope and state:** Keep pending review selection, detail and action state synchronized; a capability change must remove stale review controls and private results.
