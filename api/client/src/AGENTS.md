<a id="apiclientsrc-工作指引"></a>

# api/client/src agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [client.ts](client.ts), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Follow [parent guidance](../AGENTS.md). Before changing public requests or responses, check contracts and callers, then run API contract checks. See the [directory guide](README.md) for implementation entry points.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Type safety:** Map client calls to generated operation signatures and preserve optional, nullable and error fields through wrappers.
- **Serialization and errors:** Forward contract-defined dates, identifiers and recovery metadata without coercing them into a second wire format.
- **Verification:** Check affected consumers and generated-output freshness after changing client exports or request defaults.
