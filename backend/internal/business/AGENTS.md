<a id="backendinternalbusiness-工作指引"></a>

# backend/internal/business agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [attendance](attendance), [discussion](discussion), [enterprise](enterprise), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Domain boundaries:** Use each context's ubiquitous language and owned interfaces; shared mechanics must not merge independent business states or permissions.
- **State and transactions:** Coordinate authorization, version checks and legal transitions at the required transaction boundary; keep external effects outside retried storage callbacks.
- **Events and lifecycle:** For emitted events or jobs, define identity, schema version, duplicate handling and cleanup responsibility before adding consumers.
- **Compatibility and tests:** Preserve public errors and stored formats; verify changed rules with focused domain tests and applicable contract/emulator cases.
