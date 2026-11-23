<a id="docs-工作指引"></a>

# docs agent guidance

Inherit [root guidance](../AGENTS.md). Follow [documentation standards](standards/documentation.md); readers start at [documentation by task](README.md#按任務開始).

<a id="職責與邊界"></a>

## Responsibilities and boundaries

Define reader goals, rule ownership and verifiable evidence before choosing document locations. Every rule MUST have one authority; other documents summarize or link.

- `domains/` owns domain rules/invariants. Cross-boundary contracts own shared identity, authorization, events and API semantics; domain documents explain applications and representative cases.
- `architecture/` explains system composition/operation; `integrations/` explains external-service integration, configuration and verification. External capabilities do not establish project implementation/deployment.
- Accepted decisions belong in `decisions/`; proposals in `proposals/`. External concepts/sources may live in `references/`; adopted project rules remain with their responsible category.
- `development/`, `operations/` and `standards/` own development, operations and engineering standards respectively. Tool versions, connection settings and secrets retain their authoritative sources.

<a id="目錄配置"></a>

## Directory layout

Choose the owning category from the [directory responsibilities](README.md#目錄責任). Create a topic page when it has distinct material to maintain. Keep existing locations unless a move improves discovery or clarifies ownership.

<a id="文件與目錄命名規範"></a>

## File and directory naming

Files/directories MUST reflect actual responsibilities and business semantics with cohesion, single responsibility, clarity and consistency. Use lowercase English `kebab-case`, standard domain terminology and no vague names, redundant parent semantics/prefixes or unnecessary abbreviations. Use `README.md` for directory entry points, `AGENTS.md` for agent guidance and four-digit numbers plus `kebab-case` for ADRs. Rename only to improve semantics/boundaries, updating all references; do not refactor merely for naming uniformity.

<a id="主題子目錄"></a>

<a id="文件維護"></a>

## Documentation maintenance

`README.md` routes readers to topics; `AGENTS.md` adds local editing rules. Update category indexes/direct references for new or moved pages, and parent navigation only when entry points change. Distinguish current implementation, pending acceptance and future design; use Git for history.

<a id="圖表維護"></a>

## Diagram maintenance

Topic pages MUST own diagrams and implementation/acceptance evidence. Label current implementation, future design, development/operations guidance or current/proposed comparisons per [documentation standards](standards/documentation.md#圖表狀態與入口). Update direct references and affected category entries; preserve referenced anchors. List only frequently used cross-topic diagrams in the [project entry point](README.md#日常開發看哪張圖).

<a id="驗證"></a>

## Verification

Check relative paths, sources, anchors and indexes. Validate diagram meaning/rendering through [diagram standards](standards/documentation.md); follow the [acceptance matrix](development/tooling/testing/commands.md#變更驗收矩陣) for commands/evidence.

## Applicable engineering conventions

- **Documentation authority:** Assign each rule, contract and decision one owning topic; maintain references and anchors instead of duplicating its full text.
- **Models and compatibility:** Distinguish implemented behavior, proposed designs and migration requirements, using the same domain names as code and contracts.
- **Acceptance:** Verify changed local links and diagrams against sources; record actual evidence and limitations without declaring unexecuted tests successful.
