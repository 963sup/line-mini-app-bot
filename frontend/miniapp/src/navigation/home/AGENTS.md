<a id="frontendminiappsrcnavigationhome-工作指引"></a>

# frontend/miniapp/src/navigation/home agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **UI and routes:** Keep home actions semantically named, keyboard-accessible and aligned with the actual page entry points.
- **Identity and scope:** Preserve intended deep-link targets through login and do not confuse organization selection with resource authorization.
- **Acceptance:** Check empty membership, denied scopes and stale account data through the visible navigation controls.
