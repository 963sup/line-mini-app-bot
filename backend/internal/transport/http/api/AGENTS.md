<a id="backendinternaltransporthttpapi-工作指引"></a>

# backend/internal/transport/http/api agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [contract.go](contract.go), [routes.go](routes.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Contract wiring:** Register generated operations with their owning adapters and capability inventory; removed frontend calls require consumer-inventory review.
- **Generation and types:** Change OpenAPI sources before regenerating handlers/types; do not patch generated signatures or bypass dispatch assertions.
- **Authorization:** Keep middleware identity checks separate from resource permissions rechecked by the use case.
