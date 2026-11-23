<a id="天氣"></a>

# Weather 設計

**Implemented** · [Design](weather.md) · [Core flow](#decision-flow) · [Representative tests](../../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go) · [Module guidance](../../../backend/internal/business/weather/AGENTS.md)

<a id="領域設計"></a>

## Domain design

**Subdomain classification:** external forecasts are generic; group queries support the product.

**Language:** location is query input, providers supply forecasts and group replies present one query result.

**Context/contracts:** LINE identifies replyable group events; weather adapters query CWA. Shared location inputs do not unify attendance purpose/authorization.

**Acceptance:** group locations trigger forecasts; non-group/non-replyable events cannot trigger that reply flow.

[Implementation/tests](../../../backend/internal/business/weather/adapter/line/forecast/forecast.go) · [Strategic classification criteria](../../references/ddd/AGENTS.md)

See weather specifications for queries/LINE interaction.

[Business specification](weather.md) · [Backend weather implementation](../../../backend/internal/business/weather/AGENTS.md)

<a id="decision-flow"></a>

<a id="收到可回覆的群組位置的判定流程"></a>

## Replyable group location decision flow

```mermaid
flowchart TD
    start["Receive replyable group location"] --> gate{"Valid forecast matches location?"}
    gate -->|Yes| ok["Reply official forecast"]
    gate -->|No| stop["Report no forecast"]
```

[Entry adapter](../../../backend/internal/business/weather/adapter/cwa/forecast/cwa.go) · [Decision rules](../../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go) · [Related tests](../../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go)

[Domain guidance](../AGENTS.md)
