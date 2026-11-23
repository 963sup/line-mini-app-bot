<a id="專案基礎"></a>

# Project foundations

[domain documentation](../../../../docs/domains/project/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages construction projects, source references, inspection records, ownership and source authorization.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [inspections.go](domain/entity/inspections.go), [project.go](domain/entity/project.go), [references.go](domain/entity/references.go), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Basic specification](../../../../docs/domains/project/project.md) · [Current implementation diagram: minimal source reference structure](../../../../docs/domains/project/project.md#最小來源引用骨架)

<a id="閱讀入口"></a>

## Reading entry points

See the [basic specification](../../../../docs/domains/project/project.md) for operations, permissions, data and acceptance. Do not expand unauthorized scope. Keep core/HTTP/Firestore responsibilities separate; compose existing frontend components.

[Current implementation diagram: minimal source reference structure](../../../../docs/domains/project/project.md#最小來源引用骨架)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/project/README.md) · [decision flow](../../../../docs/domains/project/README.md#decision-flow) · [representative tests](application/reference_source_test.go)

[Parent directory guide](../README.md)
