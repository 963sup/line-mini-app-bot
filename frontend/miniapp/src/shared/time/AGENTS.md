<a id="frontendminiappsrcsharedtime-工作指引"></a>

# frontend/miniapp/src/shared/time agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [time.ts](time.ts), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Time and serialization:** Distinguish RFC 3339 UTC instants from business dates and format display values in the contract's explicit time zone.
- **Validation:** Reject ambiguous or invalid date/month inputs instead of relying on device-local parsing or silent coercion.
- **Testing:** Cover time-zone boundaries and round trips; formatting helpers must not independently decide attendance or payroll rules.
