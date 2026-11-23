<a id="摘要事件與內部工作契約"></a>

# Summary event and internal job contracts

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [worker.yaml](worker.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Summary feature](../../../docs/domains/summary/summary.md) · [Current implementation diagram: data and reliability](../../../docs/domains/summary/summary.md#資料與可靠性)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Read [existing entry contracts](README.md) and the [summary feature](../../../docs/domains/summary/summary.md). Describe message collection, summary commands and internal workers separately.
- `../internal-openapi.yaml` composes worker contracts and generates internal DTOs. Do not include them in browser clients or replace job identity with browser sessions.
- Preserve inputs, request limits, case compatibility, responses, retries, leases and failure semantics. Do not duplicate official LINE event schemas.
- Add public Summary APIs only when new feature specifications explicitly require them.

[Current implementation diagram: data and reliability](../../../docs/domains/summary/summary.md#資料與可靠性)

## Applicable engineering conventions

- **Contract semantics:** Keep internal worker identity, event/job identifiers, lease expiry and duplicate delivery semantics separate from browser sessions.
