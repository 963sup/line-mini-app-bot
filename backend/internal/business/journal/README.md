<a id="backendinternalbusinessjournal-目錄導覽"></a>

# backend/internal/business/journal directory guide

[domain documentation](../../../../docs/domains/journal/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Handles journal reads/writes. `service.go` owns business flows; `http/` connects APIs; `firestore/` stores/queries journals.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [operations.go](application/port/operations.go), [firestore](adapter/firestore), [http](adapter/http); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/journal/README.md) · [decision flow](../../../../docs/domains/journal/README.md#decision-flow) · [representative tests](domain/entity/entry_test.go)

[Existing publication operation contract and flow](../../../../docs/domains/journal/journal.md#正式群組發布的操作契約)

[Parent directory guide](../README.md)
