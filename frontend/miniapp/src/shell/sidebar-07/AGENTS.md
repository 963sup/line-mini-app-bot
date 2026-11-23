<a id="frontendminiappsrcshellsidebar-07-工作指引"></a>

# frontend/miniapp/src/shell/sidebar-07 agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [app-sidebar.tsx](app-sidebar.tsx), [nav-main.tsx](nav-main.tsx), [nav-projects.tsx](nav-projects.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Upstream compatibility:** Preserve the protected sidebar block and its public prop/event contracts; compose feature-specific behavior in the application shell.
- **UI and identity:** Use real route/account data and accessible navigation names; do not restore demo identities or fake destinations.
- **Verification:** For explicitly authorized changes, compare pinned upstream output and verify keyboard, drawer and focus behavior without blanket overwrites.
