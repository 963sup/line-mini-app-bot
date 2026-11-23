<a id="共用-api-客戶端"></a>

# Shared API client

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [package.json](package.json), [tsconfig.json](tsconfig.json), [turbo.json](turbo.json), the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Maintain the shared TypeScript client for public HTTP. Generate operation types from contracts; do not edit generated content or write parallel DTOs.
- Reuse createAPIClient and frontend session-authenticated requests for new features. Do not duplicate clients, login or retry mechanisms by capability.
- Do not add the internal Summary worker to the browser client or bundle tokens, service-account credentials or server configuration into the frontend.
- Follow parent generation and compatibility workflows for contract changes; also run relevant existing type and consumer checks for client behavior changes.

[Current implementation diagram: basic entry points](../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Types and generation:** Export generated request/response types through the shared client; repair generation sources instead of adding handwritten parallel types.
- **Identity and retries:** Keep credential, cancellation and retry behavior explicit per client; preserve the original request identity when an outcome is unknown.
- **Compatibility:** Check existing client callers when changing defaults, serialization or error propagation; never bundle internal worker credentials.
