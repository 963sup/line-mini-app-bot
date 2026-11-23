<a id="backendinternaltransporthttpbrowser-目錄導覽"></a>

# backend/internal/transport/http/browser directory guide

Provides browser routes/static files and dispatches page requests to resources.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [files.go](files.go), [routes.go](routes.go), [routes_test.go](routes_test.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)

`/miniapp/poppler` provides PDF extraction. Build checks cover the page and Go route wiring. `/_next/static/chunks/` JS/CSS use one-year immutable caching with Next-generated version hashes. HTML, other resources and private APIs retain no-cache policies.
