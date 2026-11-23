# Attendance firestore

Document mapping, queries and transactions.

<a id="實作與規則"></a>

## Implementation and rules

[Punch transactions](attendance.go), [correction transactions](attendance_correction.go), [document mapping](documents.go) and [index queries](attendance_query.go) implement storage ports. Preserve persistence paths, field names, approval evidence and transaction boundaries.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
