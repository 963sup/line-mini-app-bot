# openaiDeveloperDocs

Primary roles: Researcher, Implementer.

## When to use

Diagnose Codex agents, MCP, configuration and hooks.

## Prerequisites and configuration

Remote documentation service; network required. A website 403 does not imply MCP is unavailable.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Use `search_openai_docs` then `fetch_openai_doc`. Direct client examples appear below.

## Verification and failures

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server openaiDeveloperDocs
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).

## Direct documentation client

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server openaiDeveloperDocs --call search_openai_docs --arguments '{"query":"Codex MCP config.toml"}'
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server openaiDeveloperDocs --call fetch_openai_doc --arguments '{"url":"https://developers.openai.com/codex/mcp"}'
```

Prefer exposed host tools. The direct client is a fallback and does not register
servers in the host. Search and fetch are the only direct tool calls allowed by
this client. Report tool errors and do not treat a handshake as a successful query.

## Codex configuration policy

The configured allowlist contains search_openai_docs and fetch_openai_doc.
Agent roles inherit parent MCP servers and permissions; a role description is
not a separate filesystem sandbox. Optional server failure must be reported
without claiming host loading succeeded. See the official [MCP configuration](https://developers.openai.com/codex/mcp)
and [subagent guidance](https://developers.openai.com/codex/multi-agent).
