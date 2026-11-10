# Firestore integration

Manage the Go Firestore SDK in go.mod. Reuse an appropriately scoped client,
propagate contexts and close owned resources. Define document ownership and
validate it server-side: server SDK access requires application authorization.
Use transactions where concurrent updates require atomic read/modify/write;
transaction callbacks may retry, so keep external side effects outside them.

Configure the selected Google Cloud project, database and application identity
explicitly. Cloud Run uses the attached runtime service account and
appropriate IAM; do not assume credentials are provided by the deployment host.

The repository provides a read-only `cmd/firestorecheck` command for an explicitly
selected existing document. See [connection instructions](../development/line-prototype.md#firestore-唯讀連線).
It prints no document contents and writes no data. Application attendance persistence
and transaction boundaries remain unimplemented.

## 索引宣告

部署用索引集中於 [deploy/firestore/indexes.json](../../deploy/firestore/indexes.json)。搬移設定檔不會建立索引或切換讀取模式；套用到目標資料庫仍依發布與資料回填流程執行。
