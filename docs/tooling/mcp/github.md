# github

Primary roles: Orchestrator, Reviewer, Integrator.

## When to use

Read issues, PRs and CI state; perform authorized collaboration.

## Prerequisites and configuration

Remote service; supply GITHUB_MCP_TOKEN securely at runtime as specified in config.toml. Repository permissions and service availability are separate from token presence.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Use exposed read tools for the target repository; perform writes or pushes only within user authorization.

## Verification and failures

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server github
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).
