# Firestore 後端工具

## 唯讀連線檢查

設定 `GOOGLE_CLOUD_PROJECT`；`FIRESTORE_DATABASE_ID` 預設為 `(default)`。本機使用應用程式預設憑證（ADC），Cloud Run 使用執行服務帳戶。從儲存庫根目錄執行，將 `COLLECTION/DOCUMENT` 換成一份已存在的文件：

```bash
bash scripts/tooling/tool-exec.sh go -C backend run . check-firestore -document COLLECTION/DOCUMENT
```

[檢查程式](../../../../backend/internal/bootstrap/firestore_check.go)對指定文件執行一次 `Get`。成功代表目前身分可讀該文件；程式不印出文件內容，也不寫入資料。若收到 `NotFound`，請核對資料庫和文件路徑，讀取權限仍待確認。

## 依後端任務查閱

| 任務 | 操作或規格 |
| --- | --- |
| 在本機測試新增、更新或刪除資料 | [啟動 Firestore emulator 並執行整合測試](../testing/firestore-emulator.md) |
| 驗證交易、併發更新與重試 | [交易與一致性](../../../integrations/firestore/consistency.md) · [整合驗收流程](../testing/commands.md#firestore-integration) |
| 修改查詢、分頁或複合索引 | [查詢與索引](../../../integrations/firestore/querying.md) · [專案索引宣告](../../../../firestore.indexes.json) |
| 確認文件歸屬、TTL、清理與恢復 | [資料模型](../../../integrations/firestore/data-model.md) · [資料生命週期](../../../integrations/firestore/lifecycle.md) |

執行身分、IAM、Session 保存等主題見 [Firestore 整合索引](../../../integrations/firestore/README.md)。
