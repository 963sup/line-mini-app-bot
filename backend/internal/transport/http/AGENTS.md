<a id="backendinternaltransporthttp-工作指引"></a>

# backend/internal/transport/http agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [json.go](json.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **HTTP and errors:** Emit the shared contract error envelope, status codes and recovery metadata; never infer success from a partially written response.
- **Security and caching:** Apply existing body limits, authentication/CSRF rules and private-response cache policy at the proper boundary.
- **Reliability:** Propagate request cancellation/deadlines and verify malformed, rejected and unavailable response paths.
