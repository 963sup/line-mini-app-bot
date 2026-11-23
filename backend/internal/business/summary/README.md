<a id="backendinternalbusinesssummary-目錄導覽"></a>

# backend/internal/business/summary directory guide

[domain documentation](../../../../docs/domains/summary/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Owns LINE group message collection and summarization, including persistence, generation, scheduling and delivery.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [errors.go](application/port/errors.go), [firestore](adapter/firestore), [line](adapter/line), [messagecollection](application/messagecollection), [summarization](application/summarization); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: data and reliability](../../../../docs/domains/summary/summary.md#資料與可靠性)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../../docs/domains/summary/README.md) · [decision flow](../../../../docs/domains/summary/README.md#decision-flow) · [decision implementation](application/summarization/validation.go)

[Parent directory guide](../README.md)
