# Go delivery workflow

Eight roles are available; they are not eight mandatory execution stages.
The orchestrator can complete small work directly. Use implementer, tester and
reviewer for substantial changes; add architect/researcher/explorer only for a
specific unresolved question. Use integrator when separate patches require it.
Start with two or three independent workers, within host limits. One writer owns
each file; test fixtures must have explicit ownership too.

## Assignment contract

Provide the goal, owned files, dependencies, behavior to preserve and acceptance
commands. Return changed paths, actual commands/results, skips and open issues.
Do not repeat an investigation already answered unless new evidence conflicts.

## Acceptance matrix

| Change | Required evidence |
| --- | --- |
| Configuration/scripts | `python3 scripts/check-config.py`; relevant task smoke check |
| Go behavior | format check, focused package tests, then shared `mise run check` and `mise run build` before delivery |
| Concurrency | race tests exercising the changed concurrent behavior |
| Dependencies | inspect go.mod/go.sum; download/verify modules; vulnerability review as appropriate |
| Firestore behavior | emulator integration tests; report emulator absence explicitly |
| LINE authentication/webhook | invalid signature/token, unauthorized user and duplicate event cases |

Run tooling through `bash scripts/tool-exec.sh <command>`. Formatting can be
checked with `gofmt -l` on changed Go files; formatting changes require file
ownership. No go.mod means application verification is unavailable, not passed.
Run complete race suites once after integration instead of concurrently per role.

## Evidence limits

Graphify AST extraction can omit TOML, documentation and dynamic calls. Repomix
is a source bundle, not a correctness proof. Configuration parsing, MCP handshake,
host tool exposure and representative tool operations are separate checks.
Do not claim a fresh Cloud environment works based on an existing tool cache.

## Documentation consulted

Context7 `/openai/codex`: role configuration and delegation guidance. Relevant
upstream source: https://github.com/openai/codex/blob/main/codex-rs/core/src/agent/role.rs
and https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/handlers/multi_agents_spec.rs.
Direct retrieval of https://developers.openai.com/codex/multi-agent/ returned
HTTP 403 in this environment; Context7 excerpts are not a live host capability test.

## MCP and mise troubleshooting

For Codex configuration, agents, hooks or MCP issues, search OpenAI Docs MCP
first and fetch the relevant document. Use exposed host tools when available.
If the host does not expose them, the repository offers a direct transport client:

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server openaiDeveloperDocs --call search_openai_docs --arguments '{"query":"Codex MCP config.toml"}'
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server openaiDeveloperDocs --call fetch_openai_doc --arguments '{"url":"https://developers.openai.com/codex/mcp"}'
bash scripts/tool-exec.sh mise run doctor-openai
```

Use Context7 for version-specific Go/SDK/mise questions. Diagnose configuration,
binary availability, transport, authentication and host exposure separately.
Fix repository-owned configuration within task scope, then rerun the relevant
check. Never rewrite platform HOME/CODEX_HOME or credentials to force loading.
Missing optional GitHub credentials must not prevent standalone docs diagnosis.
Direct OpenAI Docs search and fetch both succeeded during this verification,
even though ordinary website retrieval returned 403.
