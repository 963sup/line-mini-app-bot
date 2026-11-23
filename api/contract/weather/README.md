<a id="weather--forecast-既有入口契約"></a>

# Existing Weather / Forecast entry contracts

Documents LINE weather forecast interactions.

Describes LINE forecast event entry points. The backend weather package owns forecast sources and reply logic.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Read the [local guidance](AGENTS.md) first.

<a id="相關規格"></a>

## Related specifications

[Group location weather](../../../docs/domains/weather/weather.md) · [Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

Describes existing LINE entry points without public OpenAPI or generated browser clients.
See [group location weather](../../../docs/domains/weather/weather.md) for complete product rules.

- Actor/Context: group location messages after shared webhook signature validation; retain active, reply-token and non-redelivery conditions.
- Resource/Action: query current regional forecasts by message coordinates/address without storing location.
- Authority/Policy: direct messages, ordinary text and map URLs do not trigger forecasts. Preserve CWA county/city, distance, validity window, timeout and response-size limits.
- Result: reply through LINE with official regional forecasts. Unconfigured, unsupported or upstream failure cases use existing notices; create no attendance/location records.
- Redelivery: do not repeat location-event queries. Replies are not guaranteed after failure; users may share location again.

Backend core, CWA adapter and LINE adapter live in business/weather/forecast and its cwa/ and line/ directories.
Composition is in internal/bootstrap/line.go, independent of attendance and summary storage.
There is no existing public Weather HTTP API or MINI App weather screen.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

[Parent directory guide](../README.md)
