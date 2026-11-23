<a id="frontendminiappsrcbusinessrepositorymilestones-工作指引"></a>

# frontend/miniapp/src/business/repository/milestones agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Picker.tsx](Picker.tsx), [View.tsx](View.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: milestones](../../../../../../docs/domains/repository/milestones.md#里程碑)

## Applicable engineering conventions

- **Data, scope and state:** Use server-returned milestone IDs and capability flags; distinguish an unloaded option from an invalid selection and preserve failed-page state.
