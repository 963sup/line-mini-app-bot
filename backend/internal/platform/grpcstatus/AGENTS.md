<a id="backendinternalplatformgrpcstatus-工作指引"></a>

# backend/internal/platform/grpcstatus agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [context.go](context.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Error semantics:** Translate SDK status codes into existing typed cancellation, deadline and availability outcomes without exposing provider internals.
- **Retry safety:** Keep caller cancellation distinct from retryable dependency failures and preserve wrapped-error identity.
- **Testing:** Cover changed status mappings directly; do not classify every transport failure as a safe retry.
