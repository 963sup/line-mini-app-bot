<a id="frontendminiappsrcnavigation-工作指引"></a>

# frontend/miniapp/src/navigation agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [home](home), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Routes and identifiers:** Use registered route/query conventions and encode resource identifiers without embedding credentials.
- **State and authorization:** Navigation visibility may reflect capabilities but cannot authorize a resource; preserve confirmation for dirty or uncertain writes.
- **Compatibility:** Test deep links, reload and back navigation when changing routes or selected scope.
