<a id="frontendminiappsrcidentity-工作指引"></a>

# frontend/miniapp/src/identity agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Session.tsx](Session.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: renewal transaction and frontend triggers](../../../../docs/contracts/identity.md#續期交易與前端觸發)

## Applicable engineering conventions

- **Authentication:** Keep login, session renewal and logout in the existing identity flow; backend capability data remains the authority for resource actions.
- **Cache and lifecycle:** Invalidate private client data on logout, expiry or actor change; abort old requests and prevent late responses from restoring a prior identity.
- **Errors and security:** Preserve stable authentication/recovery metadata through request helpers and never place credentials in URLs or logs.
