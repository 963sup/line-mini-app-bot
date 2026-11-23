<a id="載入本機環境啟動服務工作指引"></a>

# Local environment and service startup guidance

Inherit [parent guidance](../AGENTS.md); mise.toml owns environment variables. Keep environment loading/service startup separate.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Configuration:** Load declared non-secret settings through the existing environment entry point without duplicating defaults across scripts.
- **Lifecycle:** Manage child processes, readiness, cancellation and shutdown explicitly; a listening port alone does not prove feature acceptance.
- **Diagnostics:** Preserve original command failures and emit usable, redacted startup information.
