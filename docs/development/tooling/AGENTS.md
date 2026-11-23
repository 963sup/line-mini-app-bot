<a id="工具版本操作及診斷工作指引"></a>

# Tool versions, operations and diagnostics guidance

Inherit [parent guidance](../AGENTS.md); select topics through the [local index](./).

<a id="職責與邊界"></a>

## Responsibilities and boundaries

mise, Skills, MCP and Graphify/Repomix analysis scope/reports.
Distinguish installation, configuration, connectivity and actual operations. Coverage is not semantic correctness. Link versions/settings to their authority.

<a id="文件歸檔規則"></a>

## Filing documentation

Choose directories by primary purpose using [tool documentation standards](./). Categorize frontend, backend, API, testing and quality tools accordingly; shared development tools belong in `shared/`. Agent collaboration, Skills and general MCP settings belong in `agents/`.

Domain-specific guides stay in their domain even when called through MCP. Standard shadcn/Next DevTools paths are `frontend/shadcn.md`/`frontend/nextjs-devtools.md`, indexed by `frontend/README.md`. Other categories link the same guides.

Update category indexes, this README and old-path references when adding/moving guides. Unmoved pages retain current links.

<a id="維護與驗證"></a>

## Maintenance and verification

Update indexes/links per [documentation maintenance](../../AGENTS.md#驗證).

<a id="共用-cli-與代理接線"></a>

## Shared CLIs and agent wiring

Keep general CLI installation, configuration, operations, artifact updates/cleanup in `shared/`. `agents/` adds agent selection, MCP wiring and host verification, referencing shared commands. Check direct references, anchors and task-owned generated bundles during moves.

## Applicable engineering conventions

- **Tool structure and configuration:** File guidance by the tool's actual responsibility and link version/configuration authority instead of copying it.
- **Reproducibility and security:** Use pinned commands with explicit working directories; document required credentials without storing credential values.
- **Evidence:** Distinguish installed, configured, connected and behaviorally verified states, including timeout and recovery limitations.
