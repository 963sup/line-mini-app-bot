<a id="專案基礎"></a>

# Project foundations

Composes project selection/workspaces.

Colocate feature UI/interactions; pages compose routes and backend APIs determine permissions/results.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Inspections.tsx](Inspections.tsx), [Items.tsx](Items.tsx), [SourcePicker.tsx](SourcePicker.tsx), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Basic specification](../../../../../docs/domains/project/project.md) · [Current implementation diagram: minimal source reference structure](../../../../../docs/domains/project/project.md#最小來源引用骨架)

See the [basic specification](../../../../../docs/domains/project/project.md) for operations, permissions, data and acceptance. Do not expand unauthorized scope. Keep core/HTTP/Firestore responsibilities separate; compose existing frontend components.

[Current implementation diagram: minimal source reference structure](../../../../../docs/domains/project/project.md#最小來源引用骨架)

[Parent directory guide](../README.md)
