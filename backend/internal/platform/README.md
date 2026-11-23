<a id="backendinternalplatform-目錄導覽"></a>

# backend/internal/platform directory guide

Provides shared external-platform/data mechanisms: LINE membership queries, Firestore keys/cursors and gRPC error mapping.

Adapters share key encoding, cursors and errors. Domains own document mappings, transactions and authorization.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [line](line/README.md), [firestorekey](firestore/key), [grpcstatus](grpcstatus); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)
