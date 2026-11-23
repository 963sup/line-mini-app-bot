<a id="空間討論與留言"></a>

# Space discussions and comments

[domain documentation](../../../../docs/domains/repository/discussion/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages Repository discussion threads/comments and their operation/authorization rules.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [application service](application/service.go), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Discussion specification](../../../../docs/domains/repository/discussion/design.md) · [Current implementation diagram: minimal boundaries and verification](../../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

<a id="閱讀入口"></a>

## Reading entry points

The [discussion specification](../../../../docs/domains/repository/discussion/design.md) owns operations, permissions, data and acceptance. This module implements them without sharing Issue states.

[Current implementation diagram: minimal boundaries and verification](../../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/repository/discussion/README.md) · [decision flow](../../../../docs/domains/repository/discussion/README.md#decision-flow) · [representative tests](domain/entity/classification_test.go)

[Parent directory guide](../README.md)
