<a id="共用傳輸契約"></a>

# Shared transport contracts

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [transport.yaml](transport.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Own only genuinely shared error formats, authentication and necessary transport definitions, excluding Repository, Issue, Discussion, Project and Attendance business models.
- Shared authentication schemes describe credentials without granting group, write or review permissions. Each operation still defines authorization and failure semantics.
- Check every consumer and compatibility when changing shared definitions. Do not depend on business capabilities or composition entry points.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Contract semantics:** Keep authentication schemes and the shared error envelope transport-only; use stable error codes and optional recovery fields without duplicating domain schemas.
