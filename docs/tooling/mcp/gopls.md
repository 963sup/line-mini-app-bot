# gopls

Primary roles: Explorer, Implementer, Tester, Reviewer.

## When to use

Resolve Go symbols, references, diagnostics and affected packages.

## Prerequisites and configuration

Installed with core Go tools. Meaningful application analysis requires a valid Go module and sources. Transport success alone does not prove workspace analysis works.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Use host-exposed gopls tools after inspecting tools/list; CLI diagnostics can use `gopls check <file.go>`.

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/mcp-doctor.py --server gopls
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).
