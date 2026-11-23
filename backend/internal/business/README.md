<a id="backendinternalbusiness-目錄導覽"></a>

# backend/internal/business directory guide

Use **package by domain**. Put business rules and use cases in the responsible package; organize HTTP, LINE and Firestore adapters by domain. Collaborate through explicit cross-domain interfaces; `internal/bootstrap` composes services and routes.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [attendance](attendance), [discussion](discussion), [enterprise](enterprise), [issue](issue), [journal](journal), [object](object); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="規劃中的業務領域"></a>

## Planned business domains

[operations](operations/README.md) · [workflow](workflow/README.md) · [agent](agent/README.md) · [knowledge](knowledge/README.md)

[Parent directory guide](../README.md)
