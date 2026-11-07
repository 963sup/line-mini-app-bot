# Project documentation

These documents describe the development workflow and intended application.
Application design is a target, not evidence that Go code or deployment exists.
Read relevant pages on demand; Codex does not automatically load all docs.

- [Agent workflow](agents/workflow.md) and [acceptance](agents/acceptance.md)
- [mise](tooling/mise.md) and [seven MCP services](tooling/mcp/README.md)
- [Environment setup](development/setup.md), [Go conventions](development/go.md), [tests](development/testing.md)
- [Architecture](architecture/overview.md) and [decision records](architecture/decisions/README.md)
- [LINE](integrations/line.md) and [Firestore](integrations/firestore.md)
- [Cloud Run](operations/cloud-run.md) and [troubleshooting](operations/troubleshooting.md)

Configuration sources: `.codex/config.toml` owns roles/MCP connections;
`mise.toml` owns tool versions, non-secret environment and tasks;
`scripts/` implements tasks. Do not duplicate secrets or version pins in docs.
