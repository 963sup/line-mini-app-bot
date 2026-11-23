<a id="frontendminiappsrcbusinesssubobject-工作指引"></a>

# frontend/miniapp/src/business/subobject agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Object agent guidance](../object/AGENTS.md), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: object and subobject ownership](../../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## Applicable engineering conventions

- **Data, scope and state:** Reuse Object-owned types and editing flows for descendants; keep parent scope, branch focus and navigation locks distinct.
