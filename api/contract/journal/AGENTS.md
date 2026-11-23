<a id="apicontractjournal-工作指引"></a>

# api/contract/journal agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [journals.yaml](journals.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Follow [parent guidance](../AGENTS.md). Before changing public requests or responses, check contracts and callers, then run API contract checks. See the [directory guide](README.md) for implementation entry points.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Contract semantics:** Specify published journal visibility, source redaction, work-date representation and idempotent publication results together.
