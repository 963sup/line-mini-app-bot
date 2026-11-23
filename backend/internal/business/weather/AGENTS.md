<a id="backendinternalbusinessweather-工作指引"></a>

# backend/internal/business/weather agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [forecast](domain/entity/forecast), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="收到可回覆的群組位置的維護與驗收"></a>

## Maintaining and verifying replyable group locations

Test forecasts with group location events. Non-group or non-replyable events must not trigger the group reply flow.

When changing this rule, check the [decision flow](../../../../docs/domains/weather/README.md#decision-flow) and [representative tests](adapter/line/forecast/forecast_test.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Treat location inputs as purpose-limited data and make forecast validity, cache expiry and unavailable-result behavior explicit.
- **Cache and reliability:** Scope cached forecasts by their documented location/time semantics, bound provider calls and distinguish expired, unavailable and valid results.
- **Event acceptance:** Test replyable group events, invalid input and provider failure without introducing location history or attendance proof.
