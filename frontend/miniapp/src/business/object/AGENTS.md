<a id="frontendminiappsrcbusinessobject-工作指引"></a>

# frontend/miniapp/src/business/object agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: object and subobject ownership](../../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## Applicable engineering conventions

- **Data, scope and state:** Keep exact quantity input strings and parent/resource identities intact; display server conversion defaults and progress-reset requirements.
