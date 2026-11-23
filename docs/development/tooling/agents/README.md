# MCP tool selection

For remote data/services, discover and call tools available in this chat. Use local CLIs for uncommitted files/tests. If expected tools are absent, check [project settings](../../../../.codex/config.toml) and [loading status](#project-loading-check).

<a id="開發角色共同工作指引"></a>

## Shared development role guidance

Read [root guidance](../../../../AGENTS.md) and applicable nested instructions, following source/scope rules. Use [handoff guidance](../../workflow/handoff.md) to transfer/reuse evidence and [acceptance guidance](../testing/commands.md) to select checks. Business changes also require the [operation contract](../../../contracts/business-operations.md) and compliance flow. Run project tools through `scripts/tooling/tool-exec.sh`; report actual checks and gaps.

Exploration/review reads task-relevant sources only, distinguishes configuration from measured behavior and reports actionable findings. Recheck unchanged settings only with new evidence.

Select references by role/task. For LINE/Firestore, read [LINE capabilities](../../../integrations/line/capabilities.md)/[Firestore connectivity](../../../integrations/firestore/connection.md); locate other domains through the [documentation index](../../../README.md).

| Role | Task references |
| --- | --- |
| architect | [LINE platform architecture](../../../architecture/runtime/overview.md) · [System foundations](../../../architecture/README.md) · [graphify](graphify.md) |
| implementer | [Go development/acceptance](../../../standards/go.md) · [Acceptance matrix](../testing/commands.md) · [openaiDeveloperDocs](openai-docs.md) · [context7](context7.md) |
| explorer | [LINE platform architecture](../../../architecture/runtime/overview.md) · [graphify](graphify.md) · [gopls](../backend/gopls.md) · [repomix](repomix.md) |
| Planning/delivery | [Development workflow](../../workflow/README.md) · [Acceptance matrix](../testing/commands.md) · [github](github.md) |
| explorer (research) | [System foundations](../../../architecture/README.md) · [openaiDeveloperDocs](openai-docs.md) · [context7](context7.md) · [Repomix sources](repomix.md) |
| reviewer | [Acceptance matrix](../testing/commands.md) · [LINE platform architecture](../../../architecture/runtime/overview.md) · [repomix](repomix.md) · [github](github.md) |
| tester | [Acceptance matrix](../testing/commands.md) · [gopls](../backend/gopls.md) |

## Codex CLI

See [Codex CLI operations](codex.md) for local checkouts/Cloud tasks.

<a id="mcp-工具與角色"></a>

## MCP tools and roles

| Service | Primary roles |
| --- | --- |
| [graphify](graphify.md) | Explorer; also Architect |
| [openaiDeveloperDocs](openai-docs.md) | Explorer, Implementer |
| [context7](context7.md) | Explorer/Implementer; version-matched documentation |
| [shadcn](../frontend/shadcn.md) | Implementer; MINI App registry/integration |
| [next-devtools](../frontend/nextjs-devtools.md) | Implementer; local development runtime as needed |
| [repomix](repomix.md) | Explorer; also Reviewer |
| [mermaid](../shared/mermaid.md) | Documentation; architecture, data, flows/states; verify sources/rendering separately |
| [repomix_skills](repomix-skills.md) | Handoff owner; reusable source bundles as needed |
| [gopls](../backend/gopls.md) | Implementer; also Explorer, Tester, Reviewer |
| [github](github.md) | Exploration, documentation, planning/self-review |
| [google-developer-knowledge, gcloud](google.md) | Exploration, implementation, self-review/planning; Google docs/CLI verification |
| [cloud-run, toolbox, gke](../../../../.codex/config.toml) | Select Cloud Run, Firestore or GKE by task. Linked settings define enabled/allowed tools; verify current availability through native calls. |

The current agent selects tools by capability; this mapping is guidance, not access control. Host tool
exposure, server allowlists and user authorization determine actual access.
Use OpenAI Docs first for Codex questions, Google Developer Knowledge and installed
Google skills for Google Cloud, and Context7 for other library/version questions.
Use Graphify/gopls for structure and verify source claims; Repomix bundles sources.
TypeScript and Next.js use Context7 and version-matched official documentation;
Next.js runtime inspection may additionally use the configured next-devtools
connector. There is no dedicated TypeScript language-service MCP server.
Capability selection and writer ownership follow the [agent workflow](../../workflow/task-start.md#delegation-depth-and-role-selection).

<a id="mcp-診斷"></a>

## MCP diagnostics

[Project configuration](../../../../.codex/config.toml) defines connections, timeouts and tool allowlists. [mise.toml](../../../../mise.toml) pins tool versions. Keep token values out of documentation. For binary, network, credential and workspace failures, use [environment troubleshooting](../../environment/troubleshooting.md).

1. Run `python3 scripts/check/check-config.py` from the root to check configuration.
2. For one service, replace `<name>` with its configured name and run `bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server <name>`; for all services, run `bash scripts/tooling/tool-exec.sh mise run doctor-mcp`. Doctor checks handshakes/allowed tools. Record missing optional credentials and continue other checks.
3. Verify visibility in the actual Codex session through [host loading checks](#project-loading-check), then call a relevant tool. Record configuration, installation, handshake, host loading and operation separately.

## Host environment inheritance

Local stdio servers explicitly forward proxy and CA environment variable names
through `env_vars`. gopls additionally forwards Go module routing/privacy settings.
Values remain runtime-owned; mise still supplies workspace tool/cache paths. This
helps preserve the host network environment without embedding credentials.
Direct doctor processes inherit their caller environment and therefore do not
prove that Codex's filtered stdio launch environment or project trust is correct.
Inspect `codex mcp list` and test a host-exposed tool separately.

## Task-scoped tool use

For daily work, call tools that answer the question and reuse valid connections/results. Diagnose when tools are absent, calls fail or host settings change. Compare project settings, CLI registration and chat tools comprehensively only for new Cloud environments or explicit inventories; list disabled tools separately.

1. For a missing or failing tool, verify checkout, project trust, server
   enablement and allowlist. Prefer a native tool that matches the intended
   repository and revision. A host plugin may use a different source or index.
2. Prepare missing pinned dependencies using the existing installation task
   through `tool-exec.sh`. Run a single-server doctor when needed to identify
   startup, transport or credential failures. See the server-specific guide
   for preparation commands.
3. Use supported host rediscovery, or verify in a new session opened in the
   intended checkout. Confirm the affected server (each enabled server for a full audit) with a native read-only
   call where its allowlist permits one. If the allowlist only exposes a
   write operation, record the native tool as exposed and functional acceptance
   as pending until an authorized operation succeeds. Record missing tools, credentials
   or permissions and continue independent work. Direct transport and CLI
   checks serve as diagnostics and do not substitute for the required MCP call.

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
[doctor regression tests](../../../../scripts/test/test_mcp_doctor.py).

Local Codex sessions may disable an unrelated server without deleting its
configuration using `-c 'mcp_servers.SERVER.enabled=false'`. `enabled_tools` and
`disabled_tools` filter the configured server tools. Capability documents do not
create separate MCP connections or sandboxes. Host/plugin connections
have their own controls, so these local overrides do not reconfigure this chat.
Keep existing service capabilities unless task-specific overrides are chosen.

Configuration options and these boundaries were checked through OpenAI Docs MCP
against the [official MCP guide](https://developers.openai.com/codex/mcp) and
[subagent guide](https://learn.chatgpt.com/docs/agent-configuration/subagents) on
2026-11-14 (Asia/Taipei). This project uses one agent following capability guidance for necessary verification,
and delivers concise evidence without starting subagents.

Use repository-relative paths for Repomix file reads.

[Parent directory guide](../README.md)

## Project loading check

Open the actual Git checkout in Codex and review the host's project trust decision. From that checkout run `python3 scripts/mcp/check-host-config.py`. This compares every configured server, including disabled entries, with the effective CLI inventory. Keep project endpoints in .codex/config.toml. Installing tools does not require registering copies in the user configuration.

If CLI configuration matches, check tool discovery in the target chat and verify the required native tool with a read-only operation. A managed host may expose a separate catalog. Report that boundary when host reload or discovery is unavailable.
