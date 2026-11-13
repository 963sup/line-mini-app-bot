# MCP tool selection

| Service | Primary roles |
| --- | --- |
| [graphify](graphify.md) | Graphify；Explorer、Architect 亦可使用 |
| [openaiDeveloperDocs](openai-docs.md) | Researcher, Implementer |
| [context7](context7.md) | TypeScript、Next.js、Golang、Researcher；版本相符文件查證 |
| [shadcn](shadcn.md) | Shadcn；MINI App 元件 registry 與整合輔助 |
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
this configuration does not define dedicated TypeScript or Next.js MCP servers.
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

References: [OpenAI MCP configuration](https://developers.openai.com/codex/mcp)
and Context7 `/openai/codex` stdio environment configuration.
