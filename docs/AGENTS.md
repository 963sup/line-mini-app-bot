<a id="docs-工作指引"></a>

# docs agent guidance

Inherit [root guidance](../AGENTS.md). Follow [documentation standards](standards/documentation.md); readers start at [documentation by task](README.md#按任務開始).

<a id="職責與邊界"></a>

## Responsibilities and boundaries

Locate the owning category and authority through the [repository documentation standard](standards/documentation.md#分類命名與權威). Use the topic's own specifications and evidence to validate any change.

<a id="目錄配置"></a>

## Directory layout

Choose the owning category from the [directory responsibilities](README.md#目錄責任). Create a topic page when it has distinct material to maintain. Keep existing locations unless a move improves discovery or clarifies ownership.

<a id="文件與目錄命名規範"></a>

## File and directory naming

Follow the [repository naming rules for documentation](standards/documentation.md#分類命名與權威). When a rename clarifies ownership or meaning, update inbound references and preserve stable anchors.

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

- **Models and compatibility:** Distinguish implemented behavior, proposed designs and migration requirements, using the same domain names as code and contracts.
- **Acceptance:** Verify changed local links and diagrams against sources; record actual evidence and limitations without declaring unexecuted tests successful.
