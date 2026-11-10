# Troubleshooting

1. Run `python3 scripts/tooling/check-config.py` for syntax and document references.
2. Run Go doctor for compiler, race toolchain and writable paths.
3. Diagnose one MCP with `python3 scripts/tooling/mcp-doctor.py --server <name>` through
   tool-exec.sh. Missing optional credentials need not block another service.
4. For Codex configuration problems use [OpenAI Docs MCP](../tooling/mcp/openai-docs.md)
   search/fetch. Use [Context7](../tooling/mcp/context7.md) for version-specific APIs.
5. Repair repository-owned settings within task scope, then rerun the affected check.

Separate configured, installed, transport-ready, host-loaded and functional states.
Inspect host config/trust diagnostics when direct transport works but chat tools
are missing. Do not rewrite HOME/CODEX_HOME or platform credentials to force it.
Never print authentication headers or secret values while debugging.
