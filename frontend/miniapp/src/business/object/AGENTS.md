<a id="frontendminiappsrcbusinessobject-工作指引"></a>

# frontend/miniapp/src/business/object agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with the [repository object editor](View.tsx) and [template library](Templates.tsx). Follow the [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Object design](../../../../../docs/domains/repository/object/design.md) · [ObjectTemplate operations and permissions](../../../../../docs/domains/repository/object/template.md#操作與權限) · [OpenAPI contract](../../../../../api/contract/object/objects.yaml)

[Current implementation diagram: object and subobject ownership](../../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## Applicable engineering conventions

- **Data, scope and state:** Keep exact quantity input strings and parent/resource identities intact; display server conversion defaults and progress-reset requirements.
