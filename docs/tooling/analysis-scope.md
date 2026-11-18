# Analysis scope and evidence

`analysis-scope.toml` owns the analysis boundary. The universe is `git ls-files`
minus working-tree deletions, plus explicitly named additions and non-ignored Markdown, Mermaid and SVG files under docs/. Untracked local tools and binaries
are outside this universe; they must be explicitly added before use as evidence.
The three maintained .agents guides are included; upstream skill files stay excluded.
Every file must match a managed type/name or a documented exclusion. Unknown
types and missing sources fail classification. Exclusions record reason, impact
and owner. Paths use Python fnmatch semantics (`*` can span `/`).

Run `bash scripts/tooling/tool-exec.sh python3 scripts/analysis/analysis-scope.py` for the offline
CI classification check. The default `mise run repomix` packs this manifest, compares its paths and text with the working tree, and replaces the root XML only after verification.
Run with `--audit` for Graphify and Repomix CLI analysis
of an isolated copy of the same managed sources. Reports and raw CLI logs go to
a unique temporary directory printed by the command. The audit requires exact
Repomix file coverage and valid Graphify endpoints/IDs. Graphify coverage is
observed, not assumed: absent source nodes do not prove absence of dependencies.
Repomix coverage proves content inclusion, not semantic correctness.

The report records the commit, dirty state, source SHA-256 and contract hash.
The audit rechecks sources at the end and invalidates the run if they changed.
For reuse, follow the [evidence decision flow](../agents/workflow.md#minimal-handoff-and-evidence-reuse).
Inferred relationships, MCP host exposure and clean Cloud initialization require separate evidence.

Use `python3 scripts/analysis/analysis-scope.py --verify /absolute/path/report.json` to
reject stale sources, changed contracts or altered raw analysis artifacts.

Tool versions remain in mise.toml and the existing installation task. No tool
installation, network documentation lookup, cloud mutation or deployment occurs
in the offline classification check.
