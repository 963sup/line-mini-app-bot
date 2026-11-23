<a id="空間標籤目錄"></a>

# Space label catalog

Queries and creates Repository labels shared by Issues and Discussions.

This directory owns domain or sub-capability behavior. Add HTTP, LINE and Firestore adapters as needed; resource-owning domains control authorization and state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [labels.go](../domain/entity/labels/labels.go).

<a id="相關規格"></a>

## Related specifications

[Label specification](../../../../../docs/domains/repository/labels.md)

<a id="閱讀入口"></a>

## Reading entry points

List/create only, per the [label specification](../../../../../docs/domains/repository/labels.md); no rename/delete. Issue/Discussion author-update workflows persist classification references.

[applicable guidance](../AGENTS.md) · [parent directory guide](../README.md)
