# Attendance valueobject

Identity-free business values and validation.

<a id="實作與規則"></a>

## Implementation and rules

[Intervals](interval.go) and [locations](position.go) pass by value; validation does not mutate receivers. [Correction keys](correction_key.go) and [leave keys](leave_key.go) validate retry identifiers. To preserve evidence/serialization, intervals may represent open/invalid periods reported by calculation. Do not mutate shared slice values in place; copy and replace them.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
