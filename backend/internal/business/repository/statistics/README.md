<a id="空間統計"></a>

# Space statistics

Provides Repository statistics queries/results through `http/` APIs.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [statistics.go](../domain/entity/statistics/statistics.go).

<a id="相關規格"></a>

## Related specifications

[basic specification](../../../../../docs/domains/repository/insights.md)

<a id="閱讀入口"></a>

## Reading entry points

See the [basic specification](../../../../../docs/domains/repository/insights.md) for operations, permissions, data and acceptance. Do not expand unauthorized scope. Keep core/HTTP/Firestore responsibilities separate; compose existing frontend components.

[applicable guidance](../AGENTS.md) · [parent directory guide](../README.md)
