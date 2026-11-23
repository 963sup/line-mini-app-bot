<a id="backendinternalplatformfirestorekey-工作指引"></a>

# backend/internal/platform/firestore/key agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [cursor.go](cursor.go), [key.go](key.go), [directory guide](./), [parent guidance](../../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Identifiers and serialization:** Preserve canonical resource-key and signed-cursor encodings, including existing normalization and delimiter rules.
- **Query security:** Bind cursors to their documented resource, actor and channel scope; reject malformed or foreign cursors.
- **Compatibility:** Verify old cursor/key behavior and collision boundaries before changing formats; version an incompatible encoding explicitly.
