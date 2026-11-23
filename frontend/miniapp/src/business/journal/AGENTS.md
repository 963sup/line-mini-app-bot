<a id="frontendminiappsrcbusinessjournal-工作指引"></a>

# frontend/miniapp/src/business/journal agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [GroupCharts.tsx](GroupCharts.tsx), [WorkContext.tsx](WorkContext.tsx), [Workspace.tsx](Workspace.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Data, scope and state:** Keep repository, business date and journal link context aligned; render server progress projections without exposing inaccessible source metadata.
