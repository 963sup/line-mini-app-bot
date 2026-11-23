<a id="身分與-session-契約"></a>

# Identity and session contracts

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [identity.yaml](identity.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Session design](../../../docs/contracts/identity.md) · [Current implementation diagram: renewal transaction and frontend triggers](../../../docs/contracts/identity.md#續期交易與前端觸發)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Maintain existing MINI App public configuration, current-user and session creation, renewal and revocation contracts.
- The backend verifies trusted identity. Client input, login and profiles do not establish Repository membership or other business permissions.
- Preserve Cookie/Bearer, CSRF, errors and singleton-path compatibility. Do not expose credentials or server secrets in public DTOs.
- Read [session design](../../../docs/contracts/identity.md) before editing.

[Current implementation diagram: renewal transaction and frontend triggers](../../../docs/contracts/identity.md#續期交易與前端觸發)

## Applicable engineering conventions

- **Contract semantics:** Specify token/session expiry, CSRF requirements and revocation errors without implying that authentication grants business authorization.
