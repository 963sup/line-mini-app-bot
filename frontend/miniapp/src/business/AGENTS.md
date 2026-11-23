<a id="frontendminiappsrcbusiness-工作指引"></a>

# frontend/miniapp/src/business agent guidance

<a id="架構與責任邊界"></a>

## Architecture and responsibility boundaries

Use backend `canEdit` for content-edit controls. Test disabled capabilities while retaining separately authorized classification. Reload scoped data after account/group switches.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [attendance](attendance), [discussion](discussion), [enterprise](enterprise), [issue](issue), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Authorization and state:** Render server-provided capabilities and results without reconstructing ownership or domain transition policy in the browser.
- **Forms and recovery:** Reuse existing schema-backed fields and shared field/error mappings; preserve the original request and identity after an unknown write, and block navigation where an unresolved operation requires it.
- **Query scope:** Keep filters, cursors and projections tied to the selected authorized resource and reject late results from a prior scope.
- **UI acceptance:** Verify affected loading, empty, denied and retry states through public controls, including focus and stale-response cleanup.
