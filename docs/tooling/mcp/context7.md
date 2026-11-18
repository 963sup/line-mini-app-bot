# context7

Primary roles: Context7; TypeScript, Next.js, Golang, Researcher, Implementer and Tester may also use the service.

## When to use

Look up version-matched library, language, framework, SDK and tool documentation.

## Prerequisites and configuration

Remote service; check current schemas in tools/list. Compare documentation versions with the relevant package manifest, lockfile, go.mod or mise.toml.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Call resolve-library-id first, then query-docs with its exact library ID and a focused version-aware question. Return source URLs, documented versions and applicability limits. If the installed version is unavailable, state the mismatch rather than treating another version as verified. Do not send secrets or private source in queries.

## Anonymous fallback

Use the host Context7 plugin when available. If it reports an exhausted quota,
switch to the project `context7` MCP server and retry the lookup once. This
server is configured without an API key or credential header. Use the
anonymous server directly if the plugin is unavailable. Resolve the library ID
on the selected server before querying its documentation.

The plugin and project server are separate connections. This retry is an agent
instruction and requires the host to expose the project server. If the server
is missing from the current session, report that limitation and identify the
need for a session that loads the project configuration. If the anonymous
server also rate limits the lookup, stop retrying and report the blocker.

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/mcp-doctor.py --server context7
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../foundation/operations/troubleshooting.md).

## Source routing and evidence

Follow the [root source-verification policy](../../../AGENTS.md#mandatory-documentation-verification),
including its Codex, LINE and Google Cloud official-source routes. Context7 does
not replace those sources or define product requirements. Role handoffs follow
[the agent workflow](../../agents/workflow.md#delegation-depth-and-role-selection).
If the service or required evidence is unavailable, report the blocker and leave
the affected decision pending. A successful query does not prove application
behavior or host loading.
