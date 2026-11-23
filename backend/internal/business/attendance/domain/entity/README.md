# Attendance entity

Business objects, identity, state and invariants.

<a id="實作與規則"></a>

## Implementation and rules

[Records](record.go), [shifts](session.go), [states](state.go), [corrections](correction.go) and [leave](leave.go) store identity/lifecycle data. Domain calculations or policies validate changes.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
