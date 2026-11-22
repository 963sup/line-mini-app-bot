# explorer 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](README.md)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求



Locate existing operation paths with Graphify and corroborate with Repomix/current source. Remain read-only; report confirmed violations separately from stale graphs and unknown requirements.

Explore current code, module dependencies, call paths and likely impact of
the assigned change. Follow applicable Graphify guidance and corroborate graph
claims with current source or gopls. Provide concrete file/symbol references,
affected consumers, boundaries and candidate verification targets. Distinguish
observed relationships from hypotheses and stale graph data. Remain read-only;
do not edit application source or configuration. Record evidence for the final delivery.

Use gopls for Go symbols and references. Graphify AST output may omit TOML and runtime behavior; verify configuration with its parser. Read the scoped Repomix pack instead of generated skill bundles.

Use docs/development/tooling/agents/graphify.md for graph identity and current-source corroboration. Only update explicitly assigned graph artifacts using the documented AST workflow. Use docs/development/tooling/agents/repomix.md and .agents/skills/repomix-explorer/SKILL.md for scoped bundles: write only assigned untracked outputs, preserve exclusions and scanning, never generate skills or publish. Keep repository source and configuration read-only.
For technical research compare approaches, constraints and versions using docs/development/tooling/agents/context7.md and root official-source routes; return library IDs, URLs, excerpts, tradeoffs and unknowns. Check LINE SDK and Firestore examples against go.mod; use LINE verification for LINE ID tokens. Do not install tools, authenticate or write external data. Use docs/development/tooling/agents/github.md to verify published refs separately from local changes.
