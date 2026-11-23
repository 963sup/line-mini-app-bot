# Attendance policy

Review eligibility, retries and leave rules.

<a id="實作與規則"></a>

## Implementation and rules

[Review policy](review.go) determines independent reviewer eligibility, repeated decisions and conflicts; [leave policy](leave.go) normalizes/validates registrations. Policies do not query databases; transaction ports supply consistent facts.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
