# graphify

Primary roles: Explorer, Architect.

## When to use

Explore graph relationships before broad source reads.

## Prerequisites and configuration

Requires Graphify installed by install-mcp and the graph path configured in Codex. AST extraction may omit TOML and runtime behavior; corroborate with source/gopls.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Refresh with `mise run graphify-update`; query with `graphify query "<question>"`.

### Docs-only graph

The existing graphify-update task indexes code only; it does not extract documentation semantics.
For a separate, credential-free Markdown navigation graph:

```bash
bash scripts/tooling/tool-exec.sh mise run graphify-docs
bash scripts/tooling/tool-exec.sh graphify query "Firestore" --graph /workspace/tooling/cache/graphify-docs/graph.json
```

This graph contains ATX headings and explicit inline local links from every Markdown
file under docs/, root Markdown entry points, scripts/ guides, .agents/ top-level
guides and .github/ Markdown. Generated skill directories are excluded. It records
source locations and line coverage in coverage.json.
It does not infer semantic relationships, validate URL fragments, or automatically
identify missing requirements. Review document content to assess gaps. Generated
files stay outside the source tree and do not replace the code graph.

For synchronization, rebuild, query the affected concepts, follow related files
with path/explain, and check the source text against the accepted baseline. After
editing, rebuild and repeat the relevant queries until the scoped inconsistencies
are resolved. A valid link graph alone does not prove semantic consistency.

Full semantic extraction is a separate operation requiring a configured backend
when using the headless extract CLI. No API credentials are needed for the local
navigation graph; never store credentials in repository configuration.

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/mcp-doctor.py --server graphify
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).

Project exclusions live in `.graphifyignore` (gitignore-style). After changing
exclusions, run `graphify extract . --code-only --force --allow-dedup-shrink --out .`
through the project tooling to rescan and prune old indexed skill/example sources.
AST-only extraction does not turn Markdown plans into an implemented code graph.
