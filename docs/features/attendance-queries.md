# 出勤查詢與更正

本頁記錄現有出勤功能的資料讀取、Web 工作流程與更正重試契約。維持 Go + Next.js／React；不增加角色系統或通知。定位打卡仍沿用 [基礎出勤](attendance.md) 的本人／群組／launch 授權。

## Web 工作流程

`/miniapp` 為入口；月紀錄、更正與授權審核各自獨立，定位示範保留獨立入口。群組打卡連結仍開獨立定位打卡頁。各頁共用登入及 API 生命週期，不要求具備其他頁面的 DOM。

月份以 Asia/Taipei 月初至下月月初的半開區間計算，再轉為 UTC。跨月班次與未下班班次保留；不猜測下班時間。切換群組／月份、登出或離頁會取消讀取並使舊回應失效。頁面條件、明細與更正的班次必須一致。

## 查詢 API

所有請求仍以 LINE ID token 在伺服器驗證本人，不能提供他人的 user ID。回應不快取。新版列表／詳情不回傳精確位置或原始事件證據。

| GET 路徑（前綴 `/api/v1/attendance`） | 參數 | 回應 |
| --- | --- | --- |
| `/groups/page` | pageSize、cursor | items（id/name）、nextCursor、asOf、queryMode |
| `/monthly/summary` | group、month（YYYY-MM） | 完整月份 Calculation、count、asOf |
| `/records` | group、month、pageSize、cursor | items、nextCursor、asOf、queryMode |
| `/records/{id}` | group | 本人該群組的班次詳情／版本 |

pageSize 預設20、範圍1–100；cursor 必須與本人、channel、群組／月份與排序版本相符，無效回400 invalid_cursor。班次排序為上班時間DESC、下班時間DESC（未下班的零值排在同上班時間的已下班班次之後）、文件ID DESC，以完整排序欄位消歧義，不使用offset。沒有下一頁時 nextCursor 為空。上一頁由瀏覽器保存已使用的cursor。

月摘要計算完整相關班次，不是目前一頁的加總。既有 `/monthly`、`/groups` 與 LINE 查詢保持契約，不能靜默截成一頁。即時分頁不提供跨HTTP請求快照保證；asOf只是產生時間。核准後重新查詢，資料變更時從第一頁重新載入。

## 資料查詢切換

新寫入同交易維護 isOpen 與本人 attendanceGroups 群組投影。讀取舊schema仍相容，不能把缺少 isOpen 的舊班次誤當作不存在。

查詢索引模式僅在整個channel回填完成並明確切換後啟用；切換前新舊API皆以共用legacy reader維持可用，新API queryMode 明列 legacy。此時畫面分頁已可操作，但資料庫仍可能讀取完整歷史，不宣稱效能已改善。切換後 queryMode 為 indexed。

索引模式以已下班的時間重疊條件與未下班集合查候選班次；完整月摘要仍須讀取相關班次原始事件與更正audit證據，不能為減少讀取而省略。資料依批次讀取不等於月摘要成本固定，多range索引仍需以官方Query Explain確認實際掃描量。

部署順序：先上新雙寫程式，建立專案索引並等待ready，再執行離線回填／核對，最後啟用query marker。回填工具預設只檢查；`-apply -checkpoint /absolute/path/checkpoint.json` 僅回填，完成後另以相同命令加 `-activate` 明確確認索引ready與全部writer已升級，再切換。不能在一般HTTP請求中回填。回填可重跑、分批且以交易讀取最新班次，不能覆寫併發打卡或更正。命令細節以 `backend/cmd/attendancemigrate` 的 `--help` 為準。本次程式實作不代表遠端索引或回填已執行。

## 更正提交重試

新版Web在 `POST /api/v1/attendance/corrections` 帶一個 `Idempotency-Key` UUID。一次已確認的payload與key在結果未知時保持不變，重試不能產生新key；更改表單才建立新的提交。key不含使用者或敏感資料。

伺服器以channel、已驗證本人與key隔離回執，對固定版本的業務payload做正規化雜湊。時間以UTC正規化，breaks順序保留，server ID／CreatedAt不參與。申請與最小回執同交易建立；交易可能重跑，但不能建立第二筆申請。

- 相同key／相同payload：回原申請ID，包括後續班次已核准或版本已變動時。
- 相同key／不同payload：409 idempotency_conflict。
- 空白、無效或重複header：400 invalid_idempotency_key。
- 無header：保留舊呼叫契約，但沒有跨請求去重保證；新Web一律帶header。

提交回應 status=accepted 只表示申請已接收，不宣稱仍待審；目前狀態使用既有狀態查詢取得。申請本人與審核權限仍由後端決定，禁止自己核准；核准保留 expectedRevision／expectedEventSequence 與不可變audit驗證。回執不保存token，未確定保存政策前不擅自設定TTL。

## 驗證邊界

沿用既有檢查，不新增測試基建。Firestore emulator缺席時整合測試skip必須明列；本機編譯或純函數通過不代表遠端索引、回填、LINE裝置或正式部署已驗收。

依據：Google官方 [多range與索引](https://firebase.google.com/docs/firestore/query-data/multiple-range-fields)、[游標](https://firebase.google.com/docs/firestore/query-data/query-cursors)、[交易](https://firebase.google.com/docs/firestore/manage-data/transactions)；使用官方Developer Knowledge查得，2026-11-10核對。瀏覽器取消語意以Context7 `/mdn/content` 查證的 [AbortController](https://developer.mozilla.org/en-US/docs/Web/API/AbortController/abort) 為準；generation guard另負責拒收已完成的舊回應。

## 申請編號

新更正申請以 `github.com/google/uuid` 產生 UUID v4 作為文件與查詢／審核編號；重試從既有收據回傳已保存的編號，保留原 payload hash 衝突檢查。舊申請、收據與稽核編號不改寫，仍可查詢與審核。UUID 產生錯誤會返回失敗，不使用會 panic 的 NewString。

UUID API 已對照 [官方 v1.6.0 原始碼](https://github.com/google/uuid/blob/v1.6.0/version4.go)。Context7 團隊篩選拒絕此套件文件，因此未取得 Context7 查證；版本沿用既有 v1.6.0。
