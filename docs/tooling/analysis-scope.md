# Analysis scope and evidence

`analysis-scope.toml` owns the analysis boundary. The universe is `git ls-files`
minus working-tree deletions, plus explicitly named first-commit additions. Untracked local tools and binaries
are outside this universe; they must be explicitly added before use as evidence.
Every file must match a managed type/name or a documented exclusion. Unknown
types and missing sources fail classification. Exclusions record reason, impact
and owner. Paths use Python fnmatch semantics (`*` can span `/`).

Run `bash scripts/tooling/tool-exec.sh python3 scripts/analysis/analysis-scope.py` for the offline
CI classification check. Run with `--audit` for Graphify and Repomix CLI analysis
of an isolated copy of the same managed sources. Reports and raw CLI logs go to
a unique temporary directory printed by the command. The audit requires exact
Repomix file coverage and valid Graphify endpoints/IDs. Graphify coverage is
observed, not assumed: absent source nodes do not prove absence of dependencies.
Repomix coverage proves content inclusion, not semantic correctness.

The report records the commit, dirty state, source SHA-256 and contract hash.
Sources are checked again after analysis; changes invalidate the run. Saved
reports are evidence for their recorded sources only; compare source hashes
before reusing them. This process does not validate every inferred relationship,
MCP host exposure, or clean Cloud initialization.

Use `python3 scripts/analysis/analysis-scope.py --verify /absolute/path/report.json` to
reject stale sources, changed contracts or altered raw analysis artifacts.

Tool versions remain in mise.toml and the existing installation task. No tool
installation, network documentation lookup, cloud mutation or deployment occurs
in the offline classification check.
