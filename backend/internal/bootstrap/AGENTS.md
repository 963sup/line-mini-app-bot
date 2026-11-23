<a id="backendinternalbootstrap-工作指引"></a>

# backend/internal/bootstrap agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [attendance.go](attendance.go), [creation_line.go](../business/organization/adapter/line/creation_line.go), [group.go](group.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Dependency injection and configuration:** Assemble existing ports/adapters from validated settings; do not move business decisions into constructors.
- **Lifecycle and reliability:** Give clients and background work explicit cancellation, deadlines and cleanup ownership, including partial startup failure.
- **Observability and security:** When adding instrumentation, prefer OpenTelemetry-compatible traces/metrics and correlate structured logs; keep credentials out and metric label cardinality bounded.
