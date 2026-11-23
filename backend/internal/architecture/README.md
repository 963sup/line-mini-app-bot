<a id="backendinternalarchitecture-目錄導覽"></a>

# backend/internal/architecture directory guide

Uses architecture tests to enforce Go dependency direction and layer boundaries.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [boundaries_test.go](boundaries_test.go) and its [miniature-repository regressions](boundaries_regression_test.go). Read the [local guidance](AGENTS.md) before editing.

Tests check SDK/outer-layer dependencies and JSON/Firestore tags in all business `domain`/`application` descendants, including new packages absent from named dependency lists. Cross-context imports from every business core package require an exact entry in the scanner’s dependency map, including packages not previously registered. Same-context imports remain subject to layer rules. HTTP adapter descendants must not directly import storage or model-provider SDKs. See [DDD dependency rules](../../../docs/references/ddd/dependency-rule.md).

Adapters verify format compatibility through [journal persistence and deduplication hashes](../business/journal/adapter/firestore/payload_test.go), [organization HTTP responses](../business/organization/adapter/http/responses_test.go) and [model candidate data](../business/operations/adapter/openai/candidates_test.go). Architecture tests do not replace these contract/integration cases.

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)
