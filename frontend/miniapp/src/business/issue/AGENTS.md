<a id="frontendminiappsrcbusinessissue-工作指引"></a>

# frontend/miniapp/src/business/issue agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Data, scope and state:** Use server capabilities for content and classification controls; display Issue states without equating closure with approval.
