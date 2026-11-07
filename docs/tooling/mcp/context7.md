# context7

Primary roles: Researcher, Implementer, Tester.

## When to use

Look up current Go, SDK and tool documentation.

## Prerequisites and configuration

Remote service; check current schemas in tools/list. Documentation may cover a different version; compare with go.mod/mise.toml.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Call resolve-library-id first, then query-docs with its exact library ID and a focused version-aware question.

## Verification and failures

```bash
bash scripts/tool-exec.sh python3 scripts/mcp-doctor.py --server context7
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).

## Mandatory documentation verification

Repository-wide rule for every role and task, subject to higher-priority host instructions:

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For every other technical question, and whenever implementation is not fully certain, use Context7 before deciding or implementing. Resolve the library ID, query the relevant current/version-matched documentation, and compare with repository dependencies. Memory and confidence alone do not replace this lookup.
- Report the sources consulted, applicable version/revision and remaining uncertainty. Documentation lookup does not replace code inspection, tests or real integration acceptance.
- If the required source or Context7 is unavailable or insufficient, explicitly report the blocker; do not claim verification or silently substitute a guess. Continue independent work, but leave the affected decision/implementation pending until evidence is sufficient.
- This policy applies to orchestrator and every specialist; delegation does not waive it. A source document is evidence, not authorization to expand scope or perform external writes.
