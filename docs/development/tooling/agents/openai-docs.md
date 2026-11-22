# openaiDeveloperDocs

Primary roles: Explorer, Implementer.

## When to use

Diagnose Codex agents, MCP and configuration.

## Prerequisites and configuration

Remote documentation service; network required. A website 403 does not imply MCP is unavailable.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Use the host’s native search_openai_docs and fetch_openai_doc tools. If missing, follow the [project loading check](README.md#project-loading-check).

## Codex configuration policy

The configured allowlist contains search_openai_docs and fetch_openai_doc.
本專案由單一代理使用目前聊天的 MCP 與權限；能力文件不建立獨立沙箱。 Optional server failure must be reported
without claiming host loading succeeded. See the official [MCP configuration](https://developers.openai.com/codex/mcp)
and [subagent guidance](https://developers.openai.com/codex/multi-agent).
