<a id="出勤契約"></a>

# Attendance contracts

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [approvals.yaml](approvals.yaml), [corrections.yaml](corrections.yaml), [leave.yaml](leave.yaml), the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Attendance](../../../docs/domains/attendance/design.md) · [Queries and corrections](../../../docs/domains/attendance/queries.md)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Maintain existing punch, query, worktime, correction, review and leave operations. Preserve type ownership and dependency direction in the [capability index](../).
- Before editing, read the relevant specifications: [attendance](../../../docs/domains/attendance/design.md), [queries and corrections](../../../docs/domains/attendance/queries.md), [worktime](../../../docs/domains/attendance/worktime.md) and [leave](../../../docs/domains/attendance/leave.md).
- Preserve JSON casing, time/month formats, nullable values, pagination, request limits and Idempotency-Key semantics. Do not change public paths during naming cleanup.
- Keep login, Repository access and review permissions separate. The backend enforces personal scope, group permissions, versions and transactions. Contract validation does not replace business authorization.
- Leave registration does not imply leave approval exists. Do not advertise unimplemented operations.

[Current implementation diagram: location preflight, state and sequence](../../../docs/domains/attendance/design.md)

## Applicable engineering conventions

- **Contract semantics:** Keep instant formats, business dates, review authority and original-request retry semantics explicit for attendance operations.
