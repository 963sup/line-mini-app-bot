<a id="天氣事件契約"></a>

# Weather event contracts

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Weather feature](../../../docs/domains/weather/weather.md) · [Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Read [existing entry contracts](README.md) and the [weather feature](../../../docs/domains/weather/weather.md). These describe LINE event entry points, with no public HTTP API.
- Do not create empty OpenAPI contracts, browser clients or MINI App weather screens for directory consistency, or duplicate official LINE/CWA schemas.
- Preserve location input purpose, forecast validity, coverage and failure semantics. Do not expand into location history, attendance proof or continuous tracking.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Contract semantics:** Describe the existing event input and forecast validity without introducing a public endpoint or expanding location-data retention.
- **Event compatibility:** Follow existing provider/event schemas and receiver version support; do not add HTTP generation or a CloudEvents envelope merely for uniformity.
