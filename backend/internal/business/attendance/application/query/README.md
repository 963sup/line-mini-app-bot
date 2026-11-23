# Attendance query

Query use cases.

<a id="實作與規則"></a>

## Implementation and rules

The [monthly query](monthly.go) reads records through MonthlyReader, invokes domain calculations and maps application DTOs.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
