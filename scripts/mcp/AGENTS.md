<a id="專案維護的-mcp-註冊連線與診斷腳本工作指引"></a>

# Project MCP registration, connection and diagnostics guidance

Inherit [parent guidance](../AGENTS.md); record handshake, tool-list and actual-call results separately.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Integration contracts:** Preserve MCP framing, initialization and tool schema/version boundaries; distinguish transport probes from host tool exposure.
- **Reliability and cleanup:** Bound process/call lifetimes, close streams and terminate spawned process groups on failure or cancellation.
- **Identity and diagnostics:** Check the appropriate credential scope without printing tokens, and keep read-only probes separate from authorized writes.
