<a id="backendinternalplatformfirestorekey-目錄導覽"></a>

# backend/internal/platform/firestore/key directory guide

Provides shared Firestore document-key and pagination-cursor handling.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [cursor.go](cursor.go), [key.go](key.go), [cursor_test.go](cursor_test.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../../README.md)
