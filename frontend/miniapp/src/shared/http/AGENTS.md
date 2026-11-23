<a id="frontendminiappsrcsharedhttp-工作指引"></a>

# frontend/miniapp/src/shared/http agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [lifecycle.ts](lifecycle.ts), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Errors and retry identity:** Preserve stable API codes, versions and recovery metadata; cancellation or a lost response is not proof that a write failed.
- **Queries and cache:** Respect response cache policy, scope reads by actor/resource and discard stale results when that scope changes.
- **Reliability:** Bound requests, propagate abort signals and retry writes only according to the original request/idempotency contract.
