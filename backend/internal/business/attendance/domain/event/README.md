# Attendance event

Recorded punch facts.

<a id="實作與規則"></a>

## Implementation and rules

[Punch facts](attendance.go) record event identity, time, group, order and results; [contract versions](version.go) identify replay formats. These are recorded facts rather than punch commands; rejected attempts cannot add worktime.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
