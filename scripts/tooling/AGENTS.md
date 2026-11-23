<a id="開發工具與環境診斷工作指引"></a>

# Development tools and environment diagnostics guidance

Inherit [parent guidance](../AGENTS.md); mise.toml declares versions and tool-exec.sh starts tools.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Configuration authority:** Resolve commands and versions through the declared mise configuration and existing wrapper without hidden fallback upgrades.
- **Reliability and errors:** Preserve exit status, cancellation and bounded diagnostics across wrapper/subprocess boundaries.
- **Security and reproducibility:** Separate CLI availability from authentication and MCP exposure, recording exact versions without leaking credentials.
