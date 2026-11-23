<a id="frontendminiappsrcbusinessattendancecorrections-工作指引"></a>

# frontend/miniapp/src/business/attendance/corrections agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Requests.tsx](Requests.tsx), [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: correction submission retries](../../../../../../docs/domains/attendance/queries.md#更正提交重試)

## Applicable engineering conventions

- **Data, scope and state:** Preserve the exact correction body and retry key after an unknown submission; history pagination must not grant review authority.
