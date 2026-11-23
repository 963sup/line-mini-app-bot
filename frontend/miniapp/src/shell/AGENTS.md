<a id="frontendminiappsrcshell-工作指引"></a>

# frontend/miniapp/src/shell agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [AppFrame.tsx](AppFrame.tsx), [OrganizationSwitcher.tsx](OrganizationSwitcher.tsx), [ServiceSidebar.tsx](ServiceSidebar.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Structure and UI:** Keep application framing, navigation and identity composition separate from feature forms and protected sidebar primitives.
- **Scope and state:** Preserve intended routes and dirty/unknown-write guards across sidebar, account and organization changes.
- **Accessibility:** Maintain semantic headings, skip navigation, focus behavior and responsive layouts through existing component APIs.
