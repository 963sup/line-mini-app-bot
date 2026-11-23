<a id="apicontractobject-工作指引"></a>

# api/contract/object agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

For contract changes, follow the [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Object design](../../../docs/domains/repository/object/design.md) · [ObjectTemplate operations and permissions](../../../docs/domains/repository/object/template.md#操作與權限) · [OpenAPI contract](objects.yaml)

Before changing public requests or responses, check the [HTTP handlers](../../../backend/internal/business/object/adapter/http/) and their callers, then run the API contract checks.

[Current implementation diagram: object and subobject ownership](../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

## Applicable engineering conventions

- **Contract semantics:** Preserve exact quantity representations, object/parent identifiers and version conditions across object, subobject and history schemas.
