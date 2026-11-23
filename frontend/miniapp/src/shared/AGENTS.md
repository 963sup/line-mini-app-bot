<a id="frontendminiappsrcshared-工作指引"></a>

# frontend/miniapp/src/shared agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [FormField.tsx](FormField.tsx), [PagePanel.tsx](PagePanel.tsx), [ReadFeedback.tsx](ReadFeedback.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Component boundaries:** Share proven presentation and interaction mechanics through small typed props; do not introduce business authorization into shared UI.
- **Forms and feedback:** Keep labels, field/error associations and loading/empty/error announcements consistent across callers.
- **State ownership:** Keep derived display values out of synchronized state and clean up obsolete asynchronous work at the owning lifecycle boundary.
