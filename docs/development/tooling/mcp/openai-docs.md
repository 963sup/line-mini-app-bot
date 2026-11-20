# openaiDeveloperDocs

Primary roles: Researcher, Implementer.

## When to use

Diagnose Codex agents, MCP, configuration and hooks.

## Prerequisites and configuration

Remote documentation service; network required. A website 403 does not imply MCP is unavailable.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Use the host’s `search_openai_docs` then `fetch_openai_doc`. From the repository root,
the following task searches the official docs and registers the CLI server if missing:

```bash
bash scripts/tooling/tool-exec.sh mise run docs-openai
```

To fetch a search result by URL:

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/mcp-doctor.py --server openaiDeveloperDocs --call fetch_openai_doc --arguments '{"url":"https://developers.openai.com/codex/mcp"}'
```

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh mise run doctor-openai
```

This checks the endpoint handshake and tool allowlist without changing CLI registration.
If the CLI lacks the server, run the `docs-openai` task for OpenAI alone or `ensure-remote-mcp` for OpenAI, Context7 and GitHub. Reopen the chat to
refresh its tool inventory if needed. See [troubleshooting](../../../guides/release/troubleshooting.md)
for transport and host-loading failures.

## Codex configuration policy

The configured allowlist contains search_openai_docs and fetch_openai_doc.
Agent roles inherit parent MCP servers and permissions; a role description is
not a separate filesystem sandbox. Optional server failure must be reported
without claiming host loading succeeded. See the official [MCP configuration](https://developers.openai.com/codex/mcp)
and [subagent guidance](https://developers.openai.com/codex/multi-agent).
