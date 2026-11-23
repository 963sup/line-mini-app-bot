# openaiDeveloperDocs

Primary roles: Explorer, Implementer.

## When to use

Diagnose Codex agents, MCP and configuration.

## Prerequisites and configuration

See [shared MCP settings and diagnostics](../../environment/troubleshooting.md#設定與-mcp-診斷).

Remote documentation service; network required. A website 403 does not imply MCP is unavailable.

## Operation

Use the host’s native search_openai_docs and fetch_openai_doc tools. If missing, follow the [project loading check](../../environment/troubleshooting.md#設定與-mcp-診斷).

## Codex configuration policy

The configured allowlist contains search_openai_docs and fetch_openai_doc.
本專案由單一代理使用目前聊天的 MCP 與權限；能力文件不建立獨立沙箱。 Optional server failure must be reported
without claiming host loading succeeded. See the official [MCP configuration](https://developers.openai.com/codex/mcp)
and [subagent guidance](https://developers.openai.com/codex/multi-agent).
