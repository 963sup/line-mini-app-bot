<a id="backendinternaltransport-目錄導覽"></a>

# backend/internal/transport directory guide

Provides shared HTTP/LINE transport utilities for request/response protocols.

[HTTP](http) handles shared formats, static pages and location diagnostics. [LINE](line) handles webhook protocols/dispatch; bootstrap registers handlers in order.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [http](http), [line](line); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)
