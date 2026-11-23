<a id="frontendminiappsrcbusinessproject-工作指引"></a>

# frontend/miniapp/src/business/project agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Inspections.tsx](Inspections.tsx), [Items.tsx](Items.tsx), [SourcePicker.tsx](SourcePicker.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: minimal source reference structure](../../../../../docs/domains/project/project.md#最小來源引用骨架)

## Applicable engineering conventions

- **Data, scope and state:** Keep Project Owner, destination-copy Owner and source repository selections distinct; hiding a source or view field does not grant source access.
