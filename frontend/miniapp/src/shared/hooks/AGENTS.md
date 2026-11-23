<a id="frontendminiappsrcsharedhooks-工作指引"></a>

# frontend/miniapp/src/shared/hooks agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [use-highlighted-code.ts](use-highlighted-code.ts), [use-mobile.ts](use-mobile.ts), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Types and dependencies:** Give hooks explicit inputs/results and keep their dependency scope mechanical rather than feature-policy-specific.
- **Lifecycle and state:** Clean up listeners, timers and requests; verify remounts and late completions without duplicating derived state.
- **Component compatibility:** Preserve protected upstream hooks and public caller behavior; compose application-specific behavior outside protected files.
