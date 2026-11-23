<a id="backendinternaltransporthttp-目錄導覽"></a>

# backend/internal/transport/http directory guide

Provides shared HTTP JSON reading/writing. [api](api) wires OpenAPI; [browser](browser) serves pages; [geolocation](geolocation) handles diagnostics.

Bootstrap wraps the server with [request diagnostics](diagnostics.go), assigning a fresh server-generated `X-Request-ID` to each response. Shared service/resource error writers log 5xx failures with that ID, status, stable error code and up to eight wrapped error types. They omit raw error messages, URLs, credentials and bodies; existing error envelopes and recovery fields remain unchanged. `DEV_ERROR_DETAILS=1` retains the existing local response detail behavior and does not add raw error text to logs.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

[OpenTelemetry instrumentation](telemetry.go) exports application spans and cumulative HTTP request counts/duration histograms through the structured stdout log sink. Labels use registered route patterns, bounded methods and status codes; resource IDs and request IDs are not metric labels. Responses expose `X-Trace-ID`; W3C `traceparent` is extracted on ingress, propagated by selected dependency clients and embedded in Cloud Tasks headers so worker requests continue the producer trace. Providers flush during graceful shutdown. Raw URLs, headers, bodies, error messages and third-party span attributes are excluded. These are log exports, not a configured Cloud Trace backend, Monitoring dashboard or alert policy; cloud ingestion and retention require deployment acceptance.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [json.go](json.go), [read_json_test.go](read_json_test.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)
