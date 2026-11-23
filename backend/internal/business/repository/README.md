<a id="backendinternalbusinessrepository-目錄導覽"></a>

# backend/internal/business/repository directory guide

[domain documentation](../../../../docs/domains/repository/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages Repository ownership, queries and numbering, with Space, labels, milestones, Wiki and statistics subfeatures.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [numbering.go](domain/entity/numbering.go), [repository.go](domain/entity/repository.go), [firestore](adapter/firestore), [http](adapter/http), [labels](labels), [line](adapter/line), [milestones](milestones); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: space loading sequence](../../../../docs/domains/repository/README.md#空間載入時序)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/repository/README.md) · [decision flow](../../../../docs/domains/repository/README.md#decision-flow) · [decision implementation](domain/entity/repository.go)

[Parent directory guide](../README.md)

<a id="下層文件入口"></a>

## Child documentation entry points

- [statistics](statistics/README.md)
