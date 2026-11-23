<a id="apicontractobject-工作指引"></a>

# api/contract/object agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [objects.yaml](objects.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Follow [parent guidance](../AGENTS.md). Before changing public requests or responses, check contracts and callers, then run API contract checks. See the [directory guide](README.md) for implementation entry points.

[Current implementation diagram: object and subobject ownership](../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## Applicable engineering conventions

- **Contract semantics:** Preserve exact quantity representations, object/parent identifiers and version conditions across object, subobject and history schemas.
