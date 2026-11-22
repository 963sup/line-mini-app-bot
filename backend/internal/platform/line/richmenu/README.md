# Rich Menu 同步

提供 LINE 平台發布機制。導航內容由 transport/line/navigation 定義，bootstrap 組裝後在部署 Job 呼叫 Sync；不在服務啟動或 Webhook 中發布。

## 行為與安全邊界

先驗證所有圖片，再以內容雜湊識別選單。所有目標圖片準備完成後才更新 Alias 與預設選單；LINE Alias 更新不是原子交易，失敗可重跑收斂。內容相同不重複建立、上傳或切換。只有本次管理 Alias 或明確設定遷移對應的 managed 選單可以清理，並先確認無其他 Alias 引用、遷移使用者成功，再刪除。舊 Alias 若指向非本流程管理的選單，會在發布寫入前中止。未知預設選單查詢錯誤會中止，不擅自覆寫。

呼叫端必須對同一 Messaging API Channel 序列化發布。bootstrap 使用 Firestore 租約，所有 LINE API 呼叫共用工作 context；取消或逾時停止後續操作。對外 API 呼叫不在 Firestore transaction callback 內。

## 驗證與入口

在儲存庫根目錄執行 `bash scripts/tooling/tool-exec.sh go -C backend test -race ./internal/platform/line/richmenu ./internal/bootstrap ./internal/architecture`。測試涵蓋 staging 失敗、Alias 重跑、穩定狀態零寫入、預設查詢失敗與刪除前遷移；租約整合案例需隔離的 loopback Firestore emulator。真實 LINE 整合另驗收。

[平台入口](../README.md) · [自動部署設定](../../../../../cloudbuild.yaml) · [發布操作指南](../../../../../docs/operations/rich-menu.md)
