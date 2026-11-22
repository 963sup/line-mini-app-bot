# repomix

Primary roles: Explorer, Reviewer.

## When to use

Bundle relevant source for review and impact analysis.

## Prerequisites and configuration

安裝、CLI 打包、排除規則與參考技能更新見 [Repomix CLI](../shared/repomix.md)。Scope MCP packing explicitly.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

When the configured MCP is exposed, inspect the current tool schema before a new operation. Search the returned bundle, then corroborate critical excerpts against source. Reuse the bundle across roles when its revision, relevant working-tree changes and file coverage still match the assignment.

CLI 操作依[共用打包指引](../shared/repomix.md#打包)；host MCP exposes pack_codebase and read/search output tools.

The configured MCP runs with `--sandbox` at the repository root. Its file-tool
paths MUST be relative to that root, for example `.codex/config.toml`; absolute
paths, `~/`, `../` and `..` path segments are rejected. Reuse an existing valid
bundle or connection for related reads instead of initializing once per file.

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server repomix
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../environment/troubleshooting.md).
