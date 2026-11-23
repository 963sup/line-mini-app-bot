<a id="路由mcp-與容器驗收工作指引"></a>

# Route, MCP and container acceptance guidance

Inherit [parent guidance](../AGENTS.md); use isolated MCP test servers and local container images.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Test isolation:** Use controlled local MCP/container fixtures with explicit timeouts and process cleanup; do not depend on production credentials.
- **Contract evidence:** Assert observable route, protocol and runtime behavior for success, denial and failure without mirroring implementation details.
- **Delivery:** Report failed, passed and skipped cases separately and bind acceptance results to the tested revision.
