<a id="backendinternaltransporthttpapi-目錄導覽"></a>

# backend/internal/transport/http/api directory guide

Connects OpenAPI to Go routes and stores public/internal generated code. Start contract changes in root `api/`.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [contract.go](contract.go), [routes.go](routes.go), [generated](generated), [generatedinternal](generatedinternal); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)
