<a id="backendinternalbusinesssummary-工作指引"></a>

# backend/internal/business/summary agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [errors.go](application/port/errors.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: data and reliability](../../../../docs/domains/summary/summary.md#資料與可靠性)

<a id="驗證模型摘要的維護與驗收"></a>

## Maintaining and verifying model summaries

Check citations after model changes. Reject storage/delivery when citations are missing or outside source scope.

When changing this rule, check the [decision flow](../../../../docs/domains/summary/design.md) and [implementation](application/summarization/validation.go).

## Applicable engineering conventions

- **Domain, state and authorization:** Preserve source/citation identity, bounded model input and job expiry across retries; expired or inaccessible messages must not enter a new result.
- **Concurrency and recovery:** Coordinate job leases, completion and deduplication at the existing storage boundary; repeated delivery must not duplicate publication.
- **Data compatibility:** Preserve stored receipt/input encodings and source expiry during model changes; verify citation and lease-recovery boundaries.
