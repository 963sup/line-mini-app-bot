# Firestore 整合

<a id="firestore-integration"></a>

連線、執行身分與 client 生命週期在此維護。資料內容見[儲存模型](data-model.md)，原子提交見[交易與一致性](consistency.md)，分頁與索引見[查詢](querying.md)，過期與清除見[生命週期](lifecycle.md)。

## 執行身分與驗證

Go server SDK 使用應用身分及 IAM，不能依賴瀏覽器 Security Rules 代替後端資源授權；依[官方權限說明](https://firebase.google.com/docs/firestore/security/rules-conditions)與[業務操作模型](../../contracts/business-operations.md)核對。專案 Go Firestore SDK 依 [go.mod](../../../backend/go.mod)固定，目前為 `v1.29.2`。

每個程序明確選擇專案與 database，沿組裝入口共用適當作用域的 client，傳遞取消及期限，由建立者負責關閉。唯讀連線檢查只能證明指定讀取路徑，不能證明寫入權限、交易、索引或恢復能力。

## 既有整合說明

現行實作與設定指引

```mermaid
flowchart LR
  h["Firestore integration"]
  h --- n0["Manage the Go Firestore SDK in go.mod. Reuse<br/> an appropriately scoped client, propagate <br/>contexts and close owned resources. Define <br/>document ownership and validate it <br/>server-side: server SDK access requires <br/>application authorization. Use transactions <br/>where concurrent updates require atomic <br/>read/modify/write; transaction callbacks may<br/> retry, so keep external side effects <br/>outside them."]
  h --- n1["Configure the selected Google Cloud project,<br/> database and application identity <br/>explicitly. Cloud Run uses the attached <br/>runtime service account and appropriate IAM;<br/> do not assume credentials are provided by <br/>the deployment host."]
  h --- n2["The repository provides a read-only <br/>check-firestore application mode for an explicitly<br/> selected existing document. See connection <br/>instructions. It prints no document contents<br/> and writes no data. Application attendance <br/>persistence and transactional writes are <br/>implemented in <br/>backend/internal/business/attendance/adapter/firestore.<br/> Browser Session persistence is implemented <br/>separately; see Session and renewal design."]
  n2 --- n2l0["connection instructions"]
  click n2l0 "../line/setup.md#firestore-唯讀連線"
  n2 --- n2l1["Session and renewal design"]
  click n2l1 "../../architecture/identity/browser-session.md"
```

<a id="索引宣告"></a>

```mermaid
flowchart LR
  h["索引宣告"]
  h --- n0["部署用索引集中於 <br/>firestore.indexes.json。搬移設定檔不<br/>會建立索引或切換讀取模式；套用到目標資料庫仍<br/>依發布與資料回填流程執行。"]
  click n0 "../../../firestore.indexes.json"
  h --- n1["未來的通用資料模型見資料契約提案。"]
  click n1 "../../proposals/firestore-data-contracts.md#邏輯關係非資料庫外鍵"
```

## 官方來源與驗證範圍

2026-11-22 查閱官方 Standard Firestore 的[交易](https://firebase.google.com/docs/firestore/manage-data/transactions)、[Server SDK 權限](https://firebase.google.com/docs/firestore/security/rules-conditions)、[TTL](https://firebase.google.com/docs/firestore/ttl)及[刪除](https://firebase.google.com/docs/firestore/manage-data/delete-data)文件，對照本庫 Adapter 與 SDK 版本。這些是線上文件，未固定官方文件提交版本；本次未執行雲端設定查詢或整合測試。

[儲存索引](README.md) · [資料契約演進](../../references/ddd/model-evolution.md)
