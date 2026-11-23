<a id="frontendminiappsrcplatform-工作指引"></a>

# frontend/miniapp/src/platform agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [line](line), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Dependencies and types:** Wrap platform I/O through the existing integration boundary and expose only the types callers actually need.
- **Identity and security:** Keep provider profiles, verified sessions and business permissions distinct; do not expose server credentials.
- **Lifecycle:** Handle unsupported environments, cancellation and cleanup without duplicating platform initialization in feature pages.
