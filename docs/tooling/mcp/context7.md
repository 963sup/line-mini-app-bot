# context7

Primary roles: Researcher, Implementer, Tester.

## When to use

Look up current Go, SDK and tool documentation.

## Prerequisites and configuration

Remote service; check current schemas in tools/list. Documentation may cover a different version; compare with go.mod/mise.toml.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Call resolve-library-id first, then query-docs with its exact library ID and a focused version-aware question.

## Verification and failures

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server context7
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).
