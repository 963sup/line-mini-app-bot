<a id="frontendminiappsrcbusinessjournal-目錄導覽"></a>

# frontend/miniapp/src/business/journal directory guide

Composes journal workspaces, input/reading interactions.

Colocate feature UI/interactions; pages compose routes and backend APIs determine permissions/results.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [GroupCharts.tsx](GroupCharts.tsx), [WorkContext.tsx](WorkContext.tsx), [Workspace.tsx](Workspace.tsx), [Composer.tsx](Composer.tsx); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

- Sidebar shortcuts: Write journal → select repository → fill/publish. Feed reads the selected repository only.
- Dashboard filters dates/repositories. [GroupCharts.tsx](GroupCharts.tsx) draws one shadcn bar chart per repository with journals, showing actual today/expected tomorrow cumulative completion per work item. More pages merge into the same repository chart; expanded entries show personnel/source links.
- Charts compose official shadcn [Bar](https://ui.shadcn.com/charts/bar#charts)/[Tooltip](https://ui.shadcn.com/charts/tooltip#charts), showing full work names/percentages.

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[Parent directory guide](../README.md)

<a id="工作選項載入"></a>

## Loading work options

Within one journal draft, rows in the same repository share object/subobject/space results and in-flight requests. Failures remove entries for fresh retries; space creation invalidates options. Closing drafts/identity changes cancel/clear data without cross-draft/user storage. Subobjects still require per-object APIs rather than backend search.

Feed and composer load separately; download Composer/form dependencies only when writing opens. Charts retain lazy loading.
