# Attendance http

HTTP requests, authentication wiring and contract mapping.

<a id="實作與規則"></a>

## Implementation and rules

Handles HTTP/OpenAPI DTO mapping through [routes.go](routes.go). Identity comes from server verification; frontends cannot decide review permissions.

<a id="驗證與入口"></a>

## Verification and entry points

Inherit [module guidance](../../AGENTS.md) and run applicable checks through [module verification](../../README.md#規格與驗證). [Parent directory guide](../README.md).
