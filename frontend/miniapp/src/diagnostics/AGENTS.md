<a id="frontendminiappsrcdiagnostics-工作指引"></a>

# frontend/miniapp/src/diagnostics agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [geolocation](geolocation), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Purpose and privacy:** Keep diagnostic requests separate from business operations and avoid retaining sensitive samples beyond the documented use.
- **Errors and lifecycle:** Present permission denial, timeout and cancellation distinctly, and release watchers or requests on restart/unmount.
- **Verification:** Report browser/device limitations separately from unit or contract test results.
