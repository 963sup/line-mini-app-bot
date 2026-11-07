# repomix_skills

Primary roles: Researcher, Implementer.

## When to use

Generate a reusable skill when explicitly needed for the task.

## Prerequisites and configuration

Shares Repomix installation but has a separate tool allowlist and prompting policy. Handshake checks do not invoke generate_skill or write files.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Inspect generate_skill inputSchema, assign a destination owned by the task, then review the generated files before committing.

## Verification and failures

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server repomix_skills
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).
