<a id="儲存庫物件"></a>

# Repository objects

[domain documentation](../../../../docs/domains/repository/object/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages construction list object types, exact quantities, units and internal tracking state.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [object.go](domain/entity/object.go), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Object specification](../../../../docs/domains/repository/object/design.md) · [Current implementation diagram: object and subobject ownership](../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

<a id="閱讀入口"></a>

## Reading entry points

Object stores construction list item types, exact quantities, units and internal tracking state, separately from Issue concerns.
See the [object specification](../../../../docs/domains/repository/object/design.md) for permissions, sources and acceptance.

[Current implementation diagram: object and subobject ownership](../../../../docs/contracts/object-subobject-ownership.md#物件與子物件歸屬)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/repository/object/README.md) · [decision flow](../../../../docs/domains/repository/object/README.md#decision-flow) · [representative tests](application/progress_test.go)

[Parent directory guide](../README.md)

Objects may contain only a name; type, quantity, unit and space may be unset. Unset does not mean quantity zero or default construction work. Work with completed quantities retains exact-quantity, unit and progress-removal protections. Firestore preserves fields/schemaVersion and existing-data readability.

[Minimal ObjectTemplate workflow](../../../../docs/domains/repository/object/template.md): create immutable v1, select, apply atomically and open an independent Object tree. Drafts, revisions and legacy migration remain future work.
