# MCP tool selection

| Service | Primary roles |
| --- | --- |
| [graphify](graphify.md) | Graphify；Explorer、Architect 亦可使用 |
| [openaiDeveloperDocs](openai-docs.md) | Researcher, Implementer |
| [context7](context7.md) | Context7；TypeScript、Next.js、Golang、Researcher 亦可使用；版本相符文件查證 |
| [shadcn](shadcn.md) | Shadcn；MINI App 元件 registry 與整合輔助 |
| [next-devtools](next-devtools.md) | Next.js；按需檢查本機開發伺服器的 runtime 狀態 |
| [repomix](repomix.md) | Repomix；Explorer、Reviewer 亦可使用 |
| [repomix_skills](repomix-skills.md) | Researcher, Implementer |
| [gopls](gopls.md) | Golang；Explorer、Tester、Reviewer 亦可使用 |
| [github](github.md) | Orchestrator, Reviewer, Integrator |
| [Google Developer Knowledge / gcloud](google.md) | Researcher, Implementer, Reviewer, Integrator |

Roles share services; this mapping is guidance, not access control. Host tool
exposure, server allowlists and user authorization determine actual access.
Use OpenAI Docs first for Codex questions, Google Developer Knowledge and installed
Google skills for Google Cloud, and Context7 for other library/version questions.
Use Graphify/gopls for structure and verify source claims; Repomix bundles sources.
TypeScript and Next.js use Context7 and version-matched official documentation;
Next.js runtime inspection may additionally use the configured next-devtools
connector. There is no dedicated TypeScript language-service MCP server.
Role selection and writer ownership follow the [agent workflow](../../agents/workflow.md#delegation-depth-and-role-selection).

Configuration, installation, initialize/tools-list, host loading and functional
operations are separate verification stages. `mise run doctor-mcp` checks transport
and tool allowlists only. Missing credentials or failed requests are not PASS.
Use single-server checks when optional credentials are
unavailable. Local servers run through tool-exec.sh with workspace-owned state.

## Host environment inheritance

Local stdio servers explicitly forward proxy and CA environment variable names
through `env_vars`. gopls additionally forwards Go module routing/privacy settings.
Values remain runtime-owned; mise still supplies workspace tool/cache paths. This
helps preserve the host network environment without embedding credentials.
Direct doctor processes inherit their caller environment and therefore do not
prove that Codex's filtered stdio launch environment or project trust is correct.
Inspect `codex mcp list` and test a host-exposed tool separately.

## Task-scoped tool use

Check the current host's exposed tools once for the relevant service. Reuse the
verified inventory within the same host/session, server configuration and
allowlist; refresh when those change, a required tool is missing, or a call fails.
Do not run doctor or tools/list before every operation or independently in every
specialist. A protocol client's initial initialize/tools-list handshake is distinct
from unnecessary repeated discovery during a healthy connection.

Prefer a host-exposed tool that matches the intended repository and revision.
If absent, use the service guide's direct client or pinned CLI fallback and report
the distinction. A host plugin's graph index can differ from the local repository
graph. Loading `.codex/config.toml` cannot add tools to an already-running chat
host; inspect the host rather than repeatedly editing configuration or restarting
servers to search for nonexistent exposure.

Use only task-relevant services and bounded queries. Reuse a healthy connection
when the client supports it; keep related search/fetch or bundle/read operations
within that connection rather than reconnecting for each tool. Run a single-server
doctor for a diagnosed problem, not the full optional service suite for routine
work. On failure identify binary, arguments, transport, network or source mismatch
before retrying; do not repeatedly increase timeouts or launch unrelated servers.

The doctor bounds stdio waits and isolates each HTTP exchange in a worker that
is killed and reaped on timeout or interruption. Its existing server timeout
configuration owns the budget; this is not a hard real-time guarantee for OS
process creation. Offline regressions run in the configuration CI job; see the
[deadline and cleanup evidence](../http-deadline-convergence-20261113.md).

Local Codex sessions may disable an unrelated server without deleting its
configuration using `-c 'mcp_servers.SERVER.enabled=false'`. `enabled_tools` and
`disabled_tools` filter tools in the parent configuration; a custom role inherits
that server set and cannot create a separate MCP sandbox. Host/plugin connections
have their own controls, so these local overrides do not reconfigure this chat.
Keep existing service capabilities unless task-specific overrides are chosen.

Configuration options and these boundaries were checked through OpenAI Docs MCP
against the [official MCP guide](https://developers.openai.com/codex/mcp) and
[subagent guide](https://learn.chatgpt.com/docs/agent-configuration/subagents) on
2026-11-14 (Asia/Taipei). Independent read-heavy subproblems can use agents; their
results should be concise evidence summaries, not raw intermediate output.

### Verification snapshot

On 2026-11-14 (Asia/Taipei), a read-only Context7 registration check used one
focused local Graphify CLI query (0.813 s), then one Repomix MCP connection:
initialize 0.384 s, tools/list 0.010 s, and two file reads 0.018 s / 0.004 s.
The measured corrected run had two MCP tool calls, one CLI query and no failures;
an earlier absolute-path call was rejected by Repomix's sandbox and corrected
to repository-relative paths. These are observed operation times, not model/agent
latency or a before/after performance claim.

The source results were summarized as the workflow's six-field handoff with
revision, relevant dirty paths and source locations. This verifies direct tools
and evidence transfer; the current host has no `agent_type` selection parameter,
so native execution of all registered custom roles remains unverified. Configuration,
direct MCP operation and host-exposed tools must still be reported separately.

References: [OpenAI MCP configuration](https://developers.openai.com/codex/mcp)
and Context7 `/openai/codex` stdio environment configuration.
