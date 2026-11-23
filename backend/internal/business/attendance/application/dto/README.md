# Attendance dto

Pagination, monthly reports, boards and punch results.

<a id="實作與規則"></a>

## Implementation and rules

[Punch results](punch.go), [monthly reports](monthly.go), [pagination](page.go), [correction details](correction.go) and [boards](board.go) are application inputs/outputs without HTTP/Firestore serialization tags.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
