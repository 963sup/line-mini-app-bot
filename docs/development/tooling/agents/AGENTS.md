<a id="mcp-與工具查證指引"></a>

# MCP and tool source verification guidance

Follow [tool documentation guidance](../AGENTS.md); find tool entries in the [local index](README.md).

<a id="官方來源"></a>

## Official sources

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For Google Cloud questions, use the installed google/skills skills and Google Developer Knowledge MCP first. If unavailable, use current official Google sources and report the lookup limitation. Validate exact gcloud syntax with leaf-level help before executing commands.

For Codex questions, consult OpenAI Docs MCP first. Follow [repository verification rules](../../../../AGENTS.md#mandatory-documentation-verification) for SDKs/versions.

## gcloud sign-in routing

For gcloud login, follow [the short sign-in workflow](../shared/gcloud.md#gcloud-sign-in). Keep the current agent as the executor for the whole OAuth process; it presents the clickable URL and returns the user code to the same process. Reuse valid credentials first; otherwise provide the headless login URL immediately. Do not start browser/desktop discovery or additional specialists unless the user asks. Keep CLI login separate from ADC and resource changes.

## Applicable engineering conventions

- **Integration and authorization:** Describe the actual host-exposed tools and required resource scopes; tool visibility or a successful handshake grants no extra write authority.
- **Reliability and diagnostics:** Record bounded startup/call timeouts and distinguish configuration, protocol, credential and service failures without exposing tokens.
- **Compatibility:** Verify version-matched official sources before changing MCP, CLI or sign-in instructions; preserve their distinct credential lifecycles.
