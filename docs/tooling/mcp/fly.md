# Fly MCP

Primary roles: Orchestrator, Implementer, Reviewer, Integrator.

## Connection and prerequisites

Official setup: https://fly.io/mcp/
The HTTP MCP endpoint is https://mcp.fly.io (the setup page is not the endpoint).
It uses client-managed OAuth browser sign-in. No npm package or mise-installed
server is required. Connection and policy belong in `.codex/config.toml` under
`[mcp_servers.fly-mcp]`; credentials remain in the host's credential storage.
The server is optional and tool operations use the configured prompt policy.

## Host setup

The project already declares the connection. In a compatible Codex CLI that loads
this trusted project config, inspect `codex mcp list` then authenticate with:

```bash
codex mcp login fly-mcp
```

When registering separately in a client without the project config, the official
command is `codex mcp add fly-mcp --url https://mcp.fly.io`. Avoid duplicate entries
when project registration already works. Complete OAuth through the client;
never commit credentials or substitute a CI FLY_API_TOKEN without documented support.
Cloud/desktop hosts may expose a separate registration flow. Do not rewrite their
HOME/CODEX_HOME or authentication storage to force project configuration loading.

## Use and representative verification

Inspect exposed tool schemas first. Begin with `get_fly_context` to understand the
actual resource bindings and restrictions. For a granted app, use `get_app` and
`list_machines` to inspect identity and lifecycle status. If necessary and granted,
use `list_organizations` or `list_apps` to select the intended scope.
Do not assume account-directory access authorizes all organization resources.

The official catalog also includes Machine mutations, volumes, app secrets,
Managed Postgres and Sprites. These operations can change running services and
incur charges. Invoke them only within the current task's explicit authorization;
connection setup alone does not authorize deployment, resource creation/deletion,
secret updates, remote commands or sending feedback to Fly staff.

## Direct transport diagnosis

```bash
bash scripts/tool-exec.sh mise run doctor-fly
```

The repository doctor probes initialize/tools-list but does not implement OAuth
or load tokens from Codex credential storage. A 401 may therefore mean this direct
client needs authentication even when the host is authenticated. A 403 can also
be a network policy or service rejection; do not interpret it as successful setup.
Use host-exposed tools after OAuth for real account access verification. Do not
export OAuth credentials merely to make this unauthenticated smoke check pass.

## Troubleshooting

- Missing host tools: inspect client registration, project trust and config loading.
- Authentication required: complete OAuth in the client; inspect granted resources.
- Scope denied: request the intended resource grant through the client; do not
  silently broaden access or swap tokens.
- Network failure: check access to mcp.fly.io and OAuth destinations presented by
  the official sign-in flow; allowing only fly.io may be insufficient.
- CI deployment: see [Fly deployment](../../operations/fly.md). Its app-scoped
  deployment token and workflow are separate from MCP OAuth.

Configuration, transport, OAuth, host loading and functional operations must be
reported separately. Adding this file does not establish successful authentication.
