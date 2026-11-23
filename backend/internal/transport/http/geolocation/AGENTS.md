<a id="backendinternaltransporthttpgeolocation-工作指引"></a>

# backend/internal/transport/http/geolocation agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [geolocation.go](geolocation.go), [directory guide](./), [parent guidance](AGENTS.md).

<a id="工作方式"></a>

## Working approach

[parent guidance](../AGENTS.md)

[Current implementation diagram: attendance integration files and responsibilities](../../../../../docs/integrations/geolocation/browser-api.md#現有出勤整合檔案與責任)

## Applicable engineering conventions

- **Validation and privacy:** Validate coordinates, precision and timestamps against the location contract without treating client data as proof of business authorization.
- **HTTP semantics:** Preserve public diagnostic credential, body-limit and error semantics independently of authenticated attendance operations.
- **Testing:** Verify malformed, stale and disallowed requests; do not add location persistence or retention policy through diagnostic changes.
