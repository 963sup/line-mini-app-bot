<a id="frontendminiappsrcpages-工作指引"></a>

# frontend/miniapp/src/pages agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [_app.tsx](_app.tsx), [_document.tsx](_document.tsx), [index.tsx](index.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Structure and routing:** Keep pages as thin compositions of Shell and feature modules; use validated route parameters and existing resource identifiers.
- **Build compatibility:** Synchronize static exports, Go route registration and legacy redirects when adding or retiring a page.
- **Identity and lifecycle:** Preserve intended entry targets through login and reset resource-scoped UI on route/actor changes without discarding protected writes.
