<a id="backendinternalbusinessweather-目錄導覽"></a>

# backend/internal/business/weather directory guide

[domain documentation](../../../../docs/domains/weather/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Provides forecasts. Models live in `domain/entity/forecast/`; CWA and LINE integrations live in `adapter/cwa/forecast/` and `adapter/line/forecast/`.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [forecast](domain/entity/forecast); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/weather/README.md) · [decision flow](../../../../docs/domains/weather/README.md#decision-flow) · [representative tests](adapter/line/forecast/forecast_test.go)

[Parent directory guide](../README.md)
