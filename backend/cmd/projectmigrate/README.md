# Project 單一寫入位置遷移

此命令只處理明確指定的單一 channel／Repository 範圍。預設 dry-run，沒有隱含環境或預設資料庫，不會部署、取得新憑證或自動執行正式資料遷移。執行者必須另外具備所選資料庫權限；本次實作只在本機 emulator 驗證。

從儲存庫根目錄使用固定工具，先產生私人 manifest：

```bash
bash scripts/tooling/tool-exec.sh bash -c 'cd backend && go run ./cmd/projectmigrate -project YOUR_PROJECT -database "(default)" -channel YOUR_CHANNEL -repository YOUR_REPOSITORY_ID -manifest /absolute/private/project-migration.json'
```

manifest 包含 namespace、來源／目標路徑與 SHA-256，不含文件內容；以 mode 0600、暫存檔及 rename 保存。保留 manifest，不使用另一個 namespace 或新檔覆蓋既有證據。dry-run拒絕覆寫現有manifest，其他模式必須讀取相同manifest。

## 狀態與寫入規則

- Space ID 是 channel 內既有的穩定 `groupId`，不是 Repository ID。命令核對來源的 `id`／`groupId`，Space已存在時只接受相同身份；不存在時以既有可信群組觀察建立Space profile，並把 `repositoryUsers/{hash}/repositories/{rid}` 候選轉為獨立的 `spaceUsers/{hash}/spaces/{groupId}`。候選仍不是授權。
- 來源為 `channels/{channel}/repositories/{rid}/projects/{id}` 的完整子樹及 `projectsCreationReceipts`。目標為 `spaces/{groupId}/projects/{id}`，Project root只增加 `spaceId`、`sourceRepositoryId`，其餘欄位及items／changes保持原文；receipt保存在 `spaces/{groupId}/projectCreationScopes/{rid}/receipts/{oldReceiptId}`，保留原payloadHash及防重scope。
- 每個Project writer必須在其寫入交易內讀取 `channels/{channel}/projectStorage/{rid}`。marker不存在表示legacy；`location: legacy|space`指定唯一寫入位置，`locked: true`必須拒絕寫入。舊版不讀marker的writer尚存時不可執行；`-writers-ready`是操作者對所有writers已升級或停止的明確確認，不是命令自動驗證部署。
- apply先以交易鎖定marker，再核對原始清單/hash，逐文件create或比對相同hash續跑，絕不覆寫衝突資料。Space及候選只補不存在的資料，不覆寫較新的獨立Space觀察。完整來源與目標核對成功後，最後才原子切marker到 `space` 並解鎖。原legacy資料始終保留，沒有Project雙寫。
- 失敗／中斷保留lock及manifest，避免不完整資料被切為正式來源。重跑同一 `-apply`、namespace及manifest可以續跑。marker已切換、但本機manifest存檔中斷時，以同一migrationId恢復完成phase，不重做切換。

完成dry-run、審閱結果、確認所有writer之後，原命令加 `-apply -writers-ready`。預設總期限15分鐘，可明確調整 `-timeout`；逾時不自動解鎖。各模式不接受未指定scope的全庫操作。

## 回退與失敗恢復

原命令改加 `-rollback -writers-ready`：先鎖住space寫入，再核對legacy原文及canonical完整子樹／收據。**只有目標自apply後完全沒有新增或修改時，才切回legacy並解鎖。** 兩份資料都保留，Space profile與候選也保留；後者不授予成員權限。

若已出現新Project、Project子資料、更新版本／內容或新收據，rollback拒絕，marker保持space且鎖住；不會切回舊資料而遺失新寫入。本工具不提供「新資料copy-back」；這需要另行審閱的反向遷移。本地manifest或source漂移也會拒絕回退。

確認應繼續使用原marker位置、不要回退後，原命令可加 `-release-lock -writers-ready`。它只解除**相同manifest migrationId**持有的鎖，不改location、不覆寫／刪除資料，也不宣稱遷移完成。因衝突中止的legacy copy仍需查明原因；對錯誤canonical殘留不自動刪除，避免傷及其他資料。

## 驗證

```bash
bash scripts/tooling/tool-exec.sh bash -c 'cd backend && go test -race ./cmd/projectmigrate'
FIRESTORE_EMULATOR_HOST=127.0.0.1:8787 bash scripts/tooling/tool-exec.sh bash -c 'cd backend && go test -race ./cmd/projectmigrate -run TestMigrationLifecycle -count=1 -v'
```

一般suite檢查明確scope、dry-run預設、writer readiness、路徑／namespace隔離、來源無損轉換及manifest防篡改。emulator測試只在 `FIRESTORE_EMULATOR_HOST` 明列時執行，涵蓋完整資料／收據保存、候選可發現、cutover中斷恢復、無漂移回退、新寫入拒絕回退及衝突拒絕覆寫；不使用正式資料作測試。

交易與先讀後寫依Context7 `/googleapis/google-cloud-go` 官方transaction.go及本機Firestore v1.29.2核對；`DocumentRefs`包含缺少父文件的子集合入口，遷移遇到孤兒來源拒絕而非默默省略。Google Developer Knowledge工具未在此host提供；未宣稱實際IAM、正式資料規模、部署或live回退驗收。
