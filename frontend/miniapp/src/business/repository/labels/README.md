<a id="空間標籤目錄"></a>

# Space label catalog

Repository label listing/creation UI.

Colocate feature UI/interactions; pages compose routes and backend APIs determine permissions/results.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Picker.tsx](Picker.tsx), [View.tsx](View.tsx), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Label specification](../../../../../../docs/domains/repository/labels.md) · [Current implementation diagram: space loading](../../../../../../docs/domains/repository/README.md#空間載入時序)

List/create only per [label specifications](../../../../../../docs/domains/repository/labels.md); no rename/delete. Issue/Discussion author-update flows persist references.

`Picker` shares catalog queries/selection and changes parent drafts only; existing Issue/Discussion APIs persist them.

[Current implementation diagram: space loading sequence](../../../../../../docs/domains/repository/README.md#空間載入時序)

[Parent directory guide](../README.md)
