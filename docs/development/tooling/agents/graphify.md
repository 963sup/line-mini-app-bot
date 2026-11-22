# graphify

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

若 CLI 查詢成功但 MCP 初始化逾時，先確認本機伺服器是否回傳 JSON-RPC，再依[主機載入檢查](README.md#project-loading-check)確認目前聊天使用的服務與索引。
