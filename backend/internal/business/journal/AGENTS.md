<a id="backendinternalbusinessjournal-工作指引"></a>

# backend/internal/business/journal agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [operations.go](application/port/operations.go), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="讀取已發布日誌的維護與驗收"></a>

## Maintaining and verifying published journal reads

Readers losing source access after publication may still see work names/progress, but journals must not expose source names or IDs.

When changing this rule, check the [decision flow](../../../../docs/domains/journal/README.md#decision-flow) and [representative tests](domain/entity/entry_test.go).

[Existing publication operation contract and flow](../../../../docs/domains/journal/journal.md#正式群組發布的操作契約)

## Applicable engineering conventions

- **Domain, state and authorization:** Retain published content visibility while redacting unauthorized source data; preserve historical payload and receipt-hash encodings during model changes.
