# Attendance port

Storage, membership and transaction interfaces.

<a id="實作與規則"></a>

## Implementation and rules

[General storage](store.go), [punch capabilities](punch.go), [review transactions](review.go) and [monthly reads](monthly.go) define use-case output needs, implemented by Firestore/LINE adapters.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
