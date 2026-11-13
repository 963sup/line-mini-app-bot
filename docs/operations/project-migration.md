# Project 的 Space 保存遷移

本工具只處理一個明確 channel／來源 Repository 的舊 Project 資料，不部署、搬遷 Repository／Issue 或自動選擇正式環境。業務與授權依 [獨立 Project](../features/projects.md)。本次實作及本機 emulator 驗證不包含正式資料apply。

## 上線前提

1. 先發布或停止所有不認得 `projectStorage` lock 的舊Project writer；讀取亦須使用可辨別唯一保存位置的相容程式。不能只啟動新writer而留下舊版本同時寫入。
2. 核對明確Google project、Firestore database、Messaging API channel、sourceRepositoryId。Repository的groupId必須是有效且不衝突的Space身份；SpaceID在channel內等於groupId，不等於隨機Repository ID。
3. 每個來源使用獨立本機manifest與備份／維護紀錄。manifest保存路徑、hash、遷移ID及階段，不是授權憑證；不得提交至Git或公開分享。
4. `-writers-ready` 是操作者對部署狀態的明確確認，工具不能自動證明所有舊程序已停止。未完成此條件不得apply。

## Dry-run 與apply

透過專案固定工具執行。下列環境變數需由操作者明確設定為已確認的namespace及本機manifest；工具沒有自動選project／database的預設。

```bash
bash scripts/tooling/tool-exec.sh go -C backend run ./cmd/projectmigrate \
  -project "$MIGRATION_PROJECT" -database "$MIGRATION_DATABASE" \
  -channel "$MIGRATION_CHANNEL" -repository "$MIGRATION_REPOSITORY" \
  -manifest "$MIGRATION_MANIFEST"
```

不帶變更模式即dry-run：只讀取來源／目標，檢查identity、子樹及碰撞，建立本機manifest，沒有資料庫寫入。manifest已存在時不覆寫，使用新路徑保留先前證據。

審閱manifest、確認備份及所有writer就緒後，**另一次明確授權的環境操作**才在相同命令加入 `-apply -writers-ready`。apply先取得該遷移ID的交易鎖，再以create-only或內容完全相同的續跑處理複製，驗證完整子樹後才切換唯一寫入位置；衝突不覆寫，失敗不自動解鎖或強制切換。

範圍包括Project根、items、changes等子文件、建立收據，以及必要Space profile／候選索引。原ID、作者、版本、時間與原建立收據hash保留；只對Project根增加Space／來源識別。Space候選只供發現，不是成員授權。

## 單一writer與回復

`channels/{channelId}/projectStorage/{repositoryId}` 保存 `spaceId`、`location`、`locked`、`migrationId`。新程式於每個Project寫入交易內讀取此記錄，鎖定時拒絕變更；copy完成且驗證相符才將location切到space。不存在「同時寫舊新兩份以後再決定」的模式。

需要回復時，在相同namespace及manifest命令加入 `-rollback -writers-ready`。回復不直接把指標指回過期舊資料：必須先鎖定並核對來源及目標與manifest一致。**目前只允許未有後續漂移的回復**；canonical已有新寫入、收據或子樹改變時拒絕，不回頭覆蓋或丟失新資料。需保留這類新寫入時，另作經審閱的copy-back計畫，不能用此旗標繞過。

部分copy／驗證失敗需解除同一manifest持有的鎖時，可明確使用 `-release-lock -writers-ready`；這不刪資料、不切換保存位置，也不替失敗資料宣告驗證成功。先確認現行location及原始來源可繼續服務，保留中斷證據；異遷移ID的鎖不可釋放。

原資料與canonical副本均保留；不遞迴刪除、不重置版本、不刷新業務防重鍵。來源或目標在dry-run後改變時，重新盤點使用新的manifest，不編輯hash強迫通過。

## 驗收與移除過渡碼

隔離emulator需驗證dry-run零資料庫寫入、原ID／子樹／收據相等、同鍵重試、遷移鎖拒絕併發writer、重跑與不同migrationId衝突、來源／目標漂移拒絕、無漂移回復及新寫入後不丟資料。新舊API、Space與source權限及游標行為另依功能規格檢查。

正式apply完成後仍需確認新舊API與觀測資料相符，並保留manifest／備份至既定回復窗口結束。只有所有來源遷移完成、舊writer及API使用證據已清點後，才另行移除legacy resolver、marker、相容路由及舊資料；本工具與本文件不授權清理正式資料或部署。

[營運索引](README.md) · [Project 規格](../features/projects.md)
