<a id="frontendminiappsrcdiagnosticsgeolocation-工作指引"></a>

# frontend/miniapp/src/diagnostics/geolocation agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [View.tsx](View.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: attendance integration files and responsibilities](../../../../../docs/integrations/geolocation/browser-api.md#現有出勤整合檔案與責任)

## Applicable engineering conventions

- **Contract and types:** Send location samples through the diagnostic client using contract units and timestamps; never treat a sample as business authorization.
- **Reliability and privacy:** Keep one active watcher/request, enforce its timeout and ignore late responses without persisting location history.
- **UI states:** Present permission denial, invalid samples, cancellation and successful server validation distinctly.
