<a id="apicontractobject-目錄導覽"></a>

# api/contract/object directory guide

Defines API contracts for construction list objects and subobjects.

Maintain this capability's operations and schemas. Root specifications compose them; public formats are generated from contracts.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [objects.yaml](objects.yaml) and [local guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Objects and their child nodes share operations in [objects.yaml](objects.yaml). Read the [local guidance](AGENTS.md) before editing.

[Current implementation diagram: object and subobject ownership](../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

[Parent directory guide](../README.md)

[Minimal ObjectTemplate workflow](../../../docs/domains/repository/object/template.md): create immutable v1, select, apply atomically and open an independent Object tree. Drafts, revisions and legacy migration remain future work.
