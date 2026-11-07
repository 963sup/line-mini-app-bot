# MCP tool selection

| Service | Primary roles |
| --- | --- |
| [graphify](graphify.md) | Explorer, Architect |
| [openaiDeveloperDocs](openai-docs.md) | Researcher, Implementer |
| [context7](context7.md) | Researcher, Implementer, Tester |
| [repomix](repomix.md) | Explorer, Reviewer |
| [repomix_skills](repomix-skills.md) | Researcher, Implementer |
| [gopls](gopls.md) | Explorer, Implementer, Tester, Reviewer |
| [github](github.md) | Orchestrator, Reviewer, Integrator |
| [fly-mcp](fly.md) | Orchestrator, Implementer, Reviewer, Integrator |

Roles share services; this mapping is guidance, not access control. Host tool
exposure, server allowlists and user authorization determine actual access.
Use OpenAI Docs first for Codex questions and Context7 for library/version questions.
Use Graphify/gopls for structure and verify source claims; Repomix bundles sources.

Configuration, installation, initialize/tools-list, host loading and functional
operations are separate verification stages. `mise run doctor-mcp` checks transport
and tool allowlists only. Explicit host-OAuth services are reported as
HOST_REQUIRED with their MCP URL and host login command, not as PASS.
Optional host-only verification does not block bootstrap; real probe failures do. Use single-server checks when optional credentials are
unavailable. Local servers run through tool-exec.sh with workspace-owned state.

## Host environment inheritance

Local stdio servers explicitly forward proxy and CA environment variable names
through `env_vars`. gopls additionally forwards Go module routing/privacy settings.
Values remain runtime-owned; mise still supplies workspace tool/cache paths. This
helps preserve the host network environment without embedding credentials.
Direct doctor processes inherit their caller environment and therefore do not
prove that Codex's filtered stdio launch environment or project trust is correct.
Inspect `codex mcp list` and test a host-exposed tool separately.

References: [OpenAI MCP configuration](https://developers.openai.com/codex/mcp)
and Context7 `/openai/codex` stdio environment configuration.
