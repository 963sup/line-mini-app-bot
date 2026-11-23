<a id="backendinternalidentityhttp-工作指引"></a>

# backend/internal/identity/http agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [identity.go](identity.go), [session.go](session.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: renewal transaction and frontend triggers](../../../../docs/contracts/identity.md#續期交易與前端觸發)

## Applicable engineering conventions

- **Authentication and security:** Preserve Cookie/Bearer and CSRF semantics, stable authentication errors and no-store behavior for private session responses.
- **Time and identifiers:** Preserve token/session identifier formats and expiry precision; distinguish UTC instants from local business dates.
- **Reliability and tests:** Bound verification calls and test invalid, expired, revoked and cancelled cases at the changed boundary.
