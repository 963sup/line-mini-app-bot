# graphify

Primary roles: Explorer, Architect.

## When to use

Explore graph relationships before broad source reads.

## Prerequisites and configuration

Requires Graphify installed by install-mcp and the graph path configured in Codex. AST extraction may omit TOML and runtime behavior; corroborate with source/gopls.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Prefer the configured Graphify MCP when exposed. Use the current host inventory and configured allowlist. When using another host, verify its repository and indexed revision before comparing its results with the local graph.

CLI 建圖、排除規則與更新方式見 [Graphify CLI](../shared/graphify.md)。

### Docs-only graph

文件導航圖的建立與限制見 [共用操作](../shared/graphify.md#docs-only-graph)。

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server graphify
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../environment/troubleshooting.md).

2026-11-14 (Asia/Taipei): the configured local command and graph path were valid,
and CLI queries succeeded. Local MCP initialize did not return a JSON-RPC reply;
a bounded minimal server using the same installed MCP SDK also stalled without
importing Graphify. The problem is bounded to that environment's SDK stdio/dispatch
layer, not proven to be graph size, CLI arguments or a slow graph load. No verified
dependency fix was available, so pins and timeouts were left unchanged. Reuse this
failure evidence for the unchanged task to avoid repeating the same timeout in each role. For missing native tools follow the root exposure procedure. The host's separate cloud Graphify
index is not evidence that this local MCP or unpushed sources were loaded.
