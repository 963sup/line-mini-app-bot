<a id="backendinternaltransport-工作指引"></a>

# backend/internal/transport agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [http](http), [line](line), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Boundaries:** Translate verified requests/events into use-case calls; transport adapters do not own domain authorization or state transitions.
- **Errors and data:** Preserve the owning contract's serialization, identifiers and stable failure semantics across entry points.
- **Diagnostics:** Correlate request/event handling using structured metadata while excluding credentials and sensitive bodies.
