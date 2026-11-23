<a id="backendinternalplatform-工作指引"></a>

# backend/internal/platform agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [line](line/README.md), [firestorekey](firestore/key), [grpcstatus](grpcstatus), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Dependencies:** Keep platform helpers limited to shared mechanisms and inject provider clients rather than embedding business policy.
- **Reliability:** Propagate cancellation and typed failures, bound call deadlines and retry backoff, and assess persistent-failure isolation before introducing circuit breakers.
- **Diagnostics:** Propagate correlation information and classify failures without logging secrets or full sensitive payloads.
