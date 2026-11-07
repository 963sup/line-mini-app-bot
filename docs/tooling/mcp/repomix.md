# repomix

Primary roles: Explorer, Reviewer.

## When to use

Bundle relevant source for review and impact analysis.

## Prerequisites and configuration

Node and Repomix installed by install-mcp. .repomixignore owns project exclusions; the mise repomix task selects XML output and parsable style; explicitly scope MCP packing too. Security scanning is not a proof that no secrets exist.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Run `mise run repomix`, then read the generated XML; host MCP exposes pack_codebase and read/search output tools.

## Verification and failures

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server repomix
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).

Project exclusions use gitignore syntax in `.repomixignore`; keep project docs,
agent configuration and application source, while excluding upstream skills and
agent-specific copies. The JSON config has been removed. Gitignore/default exclusions and security scanning remain enabled by Repomix defaults; do not pass disabling flags.
