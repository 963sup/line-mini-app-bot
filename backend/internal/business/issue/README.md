<a id="空間問題"></a>

# Space issues

[domain documentation](../../../../docs/domains/repository/issue/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages Repository issues, their states and Issue-specific operation/authorization rules.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [issue.go](domain/entity/issue.go), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Issue specification](../../../../docs/domains/repository/issue/design.md) · [Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="閱讀入口"></a>

## Reading entry points

The [Issue specification](../../../../docs/domains/repository/issue/design.md) owns permissions, states, APIs, data and acceptance; this module contains corresponding implementation only.

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/repository/issue/README.md) · [decision flow](../../../../docs/domains/repository/issue/README.md#decision-flow) · [representative tests](domain/entity/issue_test.go)

[Parent directory guide](../README.md)
