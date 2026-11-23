<a id="attendance-套件分工"></a>

# Attendance package responsibilities

[domain documentation](../../../../docs/domains/attendance/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages punches, records, correction reviews, leave and queries. Adapter/application dependencies point toward the domain; `internal/bootstrap` composes external implementations.

<a id="現行結構"></a>

## Current structure

```text
attendance/
├── domain/
│   ├── entity/
│   ├── valueobject/
│   ├── service/
│   ├── event/
│   └── policy/
├── application/
│   ├── command/
│   ├── query/
│   ├── port/
│   └── dto/
└── adapter/
    ├── http/
    ├── line/
    ├── firestore/
    ├── messaging/
    └── scheduler/
```

<a id="閱讀入口"></a>

## Reading entry points

The [domain model](domain/README.md) separates entities, values, calculations, events and policies. The [application layer](application/README.md) separates writes, queries, output ports and DTOs; [adapters](adapter/README.md) handle protocols and persistence. `messaging`/`scheduler` are reserved entry points without queue/scheduler implementations.

[command](application/command/review.go) coordinates correction review; the [transaction port](application/port/review.go) supplies current facts. Firestore runs use cases, verifies versions and saves approval evidence in one transaction, reevaluating on retries. HTTP/LINE share the [monthly query](application/query/monthly.go).

<a id="規格與驗證"></a>

## Specifications and verification

[Attendance specification](../../../../docs/domains/attendance/README.md) · [Agent guidance](AGENTS.md) · [Acceptance matrix](../../../../docs/development/tooling/testing/commands.md#變更驗收矩陣). Run Go checks from the root through `bash scripts/tooling/tool-exec.sh mise run check`; Firestore integration tests also require `FIRESTORE_EMULATOR_HOST`.

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/attendance/README.md) · [decision flow](../../../../docs/domains/attendance/README.md#decision-flow) · [representative tests](adapter/firestore/attendance_integration_test.go)

[Parent directory guide](../README.md)
