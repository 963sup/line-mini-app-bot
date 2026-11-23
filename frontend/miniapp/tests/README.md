<a id="frontendminiapptests-目錄導覽"></a>

# frontend/miniapp/tests directory guide

MINI App interaction, request lifecycle/data tests and DOM setup.

Owns test environments/observable behavior checks; application code lives in src.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [dom-setup.ts](dom-setup.ts), [session-controls.tsx](session-controls.tsx), [sidebar-dom.ts](sidebar-dom.ts), [lifecycle.test.tsx](lifecycle.test.tsx), [object-spaces.test.tsx](object-spaces.test.tsx), [objects.test.tsx](objects.test.tsx); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[Parent directory guide](../README.md)
