# Attendance line

LINE events and attendance message presentation.

<a id="實作與規則"></a>

## Implementation and rules

[Punch entry points](attendance.go) and [monthly presentation](attendance_history.go) isolate LINE protocols. Group-name formatting lives in [shared presentation](../presentation/group.go).

Bootstrap injects [platform/line](../../../../platform/line/README.md) into the existing application Membership port. HTTP adapters distinguish allowed, denied and temporarily unverifiable membership.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
