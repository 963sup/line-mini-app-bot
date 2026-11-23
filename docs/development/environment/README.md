<a id="開發環境"></a>

# Development environment

<a id="依任務選擇能力"></a>

## Select capabilities by task

The current agent selects and reads applicable guidance below for each chat. [Root AGENTS.md](../../../AGENTS.md#從這裡開始) owns mandatory reading. Tasks may use multiple capabilities without reading all guides or starting subagents.

| Task | Guidance |
| --- | --- |
| Planning, scope, risk/delivery | [Planning](../workflow/capabilities/planning.md) |
| Source, dependencies/official verification | [Exploration](../workflow/capabilities/explorer.md) |
| Models, architecture/transactions | [Design](../workflow/capabilities/architect.md) |
| Diagnose and fix | [Implementation](../workflow/capabilities/implementer.md) → [Verify fixes](../workflow/capabilities/tester.md) |
| Read-only diff/risk review | [Self-review](../workflow/capabilities/reviewer.md) |
| Regression, contract/integration acceptance | [Testing](../workflow/capabilities/tester.md) |
| Markdown, Mermaid/authorized Wiki work | [Documentation](../workflow/capabilities/documentation.md) |

See [capabilities](../workflow/capabilities/README.md) for content/stage boundaries. Reading guidance does not require installation, login or application startup; perform environment operations only as needed.

<a id="依環境選擇操作"></a>

## Select operations by environment

Run commands from the root. Linux x64/arm64 requires Bash, curl, tar, sha256sum and Python 3.11+.

| Scenario | Entry point |
| --- | --- |
| Prepare/update published Codex Cloud environments | [Cloud installation/startup/acceptance](codex-cloud.md) |
| Local development | `bash scripts/bootstrap.sh core`, then [start applications](#啟動應用) as needed |
| Local MCP/cloud tools | [Bootstrap modes](../tooling/shared/toolchain-setup.md#一般開發與-mcp-安裝) |
| Configuration/startup/staged acceptance | [Environment verification](verification.md) |
| Missing tools/dependency/authentication failures | [Troubleshooting](troubleshooting.md) |

<a id="啟動應用"></a>

## Start applications

Prepare locked dependencies and required settings. Run commands separately to isolate Next.js/Go failures. Pages requiring APIs use Go on default port 8080.

| Verification target | Startup command | Readiness evidence |
| --- | --- | --- |
| Next.js pages | `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp dev` | Open the changed page on port 3000 |
| Go HTTP/browser APIs | `bash scripts/tooling/tool-exec.sh mise run dev` | Check startup logs, request routes and verify responses |
| Go/static-page integration | Generate assets with `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp build`, then start Go | Open pages through Go and check assets/APIs |

Read terminal output for each process; stop foreground processes with Ctrl+C. Sources: [Go entry](../../../scripts/dev/run.sh) · [Frontend tasks](../../../frontend/miniapp/package.json) · [Development API forwarding](../../../frontend/miniapp/next.config.ts).

[Development handbook](../README.md)
