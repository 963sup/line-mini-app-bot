# gopls

Primary roles: Explorer, Implementer, Tester, Reviewer.

## When to use

Resolve Go symbols, references, diagnostics and affected packages.

## Prerequisites and configuration

Installed with core Go tools. Meaningful application analysis requires a valid Go module and sources. Transport success alone does not prove workspace analysis works.

## Operation

Reuse the verified host inventory; inspect tools/list when it is missing or stale. Use the exposed workspace, search, file context, package API, symbol references and diagnostics tools as needed. Use go_vulncheck only when available and relevant. Detached gopls reads saved files; CLI diagnostics can use `gopls check <file.go>`. If the native tools are missing, follow the [host loading check](../../environment/troubleshooting.md#設定與-mcp-診斷).

## Verification and failures

Follow [MCP diagnostics](../../environment/troubleshooting.md#設定與-mcp-診斷) with the service name `gopls`.
