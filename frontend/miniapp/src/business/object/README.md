<a id="儲存庫物件"></a>

# Repository objects

Construction object lists, details/editing.

Colocate feature UI/interactions; pages compose routes and backend APIs determine permissions/results.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Object specification](../../../../../docs/domains/repository/object/design.md)

Object lists/details, manual creation and creator editing follow [object specifications](../../../../../docs/domains/repository/object/design.md) for fields, permissions/state.

`View` resets by repository/identity generation. Unknown creation results retain original keys/content; parent navigation preserves forms until resolved. Reuse typed clients, cancellation and official components.

[Current implementation diagram: object and subobject ownership](../../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

[Parent directory guide](../README.md)

Creation asks only for names without space loading; add details later. Unclassified objects hide progress/state. Candidate review preserves source evidence; creation never auto-applies unconfirmed quantities/classification.

[Minimal ObjectTemplate workflow](../../../../../docs/domains/repository/object/template.md): create immutable v1, select, apply atomically and open an independent Object tree. Drafts, revisions and legacy migration remain future work.
