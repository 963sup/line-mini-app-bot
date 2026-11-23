<a id="backendinternalbusinessattendance-工作指引"></a>

# backend/internal/business/attendance agent guidance

<a id="用途與範圍"></a>

## Purpose and scope

Manages punches, records, corrections, leave, reviews and attendance boards. Domain/application packages follow responsibilities in the [module guide](README.md#現行結構).

<a id="架構與責任邊界"></a>

## Architecture and responsibility boundaries

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [errors.go](application/port/errors.go), [store.go](application/port/store.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: location preflight, state and sequence](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序)

<a id="分層約束"></a>

## Layering constraints

Uses domain/entity, valueobject, service, event, policy and application/command, query, port, dto. Put new Go files in the corresponding responsibility package. Do not restore mixed feature packages or bypass separation through aliases. `internal/architecture/boundaries_test.go` checks allowed dependencies. Domains contain no query pagination or external serialization tags. `adapter/messaging` and `adapter/scheduler` are reserved entry points, not implementations.

<a id="群組打卡的維護與驗收"></a>

## Maintaining and verifying group punches

Verify separate shifts for the same person across groups and cover shift states that reject punches.

When changing this rule, check the [decision flow](../../../../docs/domains/attendance/README.md#decision-flow) and [representative tests](adapter/firestore/attendance_integration_test.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Use the attendance contract's business time zone and domain transitions for punches, corrections and reviews; adapters must not independently recompute worktime.
