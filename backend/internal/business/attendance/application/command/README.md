# Attendance command

Write use cases.

<a id="實作與規則"></a>

## Implementation and rules

[Correction review](review.go) authorizes and evaluates domain rules within port-provided transaction views, storing approval evidence and decisions. Do not read versions outside transactions and overwrite unconditionally.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
