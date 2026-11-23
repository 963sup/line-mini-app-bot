# Attendance service

Cross-object calculation, corrections and evidence checks.

<a id="實作與規則"></a>

## Implementation and rules

[Worktime calculation](worktime.go), [transitions](transition.go), [event sequence checks](event_stream.go), [correction application](correction.go), [approval evidence checks](approval.go) and [monthly calculation](monthly.go) are pure Go domain operations. Results do not own API pagination or presentation.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
