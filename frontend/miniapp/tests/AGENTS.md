<a id="frontendminiapptests-工作指引"></a>

# frontend/miniapp/tests agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [dom-setup.ts](dom-setup.ts), [session-controls.tsx](session-controls.tsx), [sidebar-dom.ts](sidebar-dom.ts), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Test boundaries:** Exercise observable behavior through public controls and contract-shaped fixtures rather than reproducing component internals.
- **Isolation and lifecycle:** Restore browser mocks, sessions and mounted roots between cases; explicitly verify denial, cancellation and unknown-write retries where relevant.
- **Acceptance evidence:** Keep unit/DOM, contract, browser and real-device claims distinct and bind results to the source revision tested.
