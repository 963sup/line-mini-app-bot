# 資料模型

資料模型由各 bounded context 的業務語意驅動；Firestore adapter 擁有保存格式與相容轉譯。公開 API、Domain model 與 document schema 不必一一相同。

## 文件設計

- 文件必須有明確 owner、作用域、ID、建立／更新時間與必要版本條件。
- 時間保存為 UTC timestamp；業務日期另保存其語意與時區。
- optional 欄位一致使用省略、`null` 或明確零值，不混用。
- ID 必須驗證，禁止 slash 與路徑控制字元進入動態 document path。
- 大型內容、檔案與不可查詢 binary 優先放 Cloud Storage，Firestore 保存 reference 與 metadata。

## 關係與唯一性

Firestore 不提供外鍵。引用需定義來源不存在、停用或刪除時的行為；唯一性在 transaction 內以可競爭的占用文件或其他原子機制保護，不以先查後寫保證。

## 交易與一致性

先從不變條件推導 transaction 的讀寫集合。callback 內只做可安全重跑的資料操作；版本、授權與狀態必須在提交前重新確認。外部 API 不屬於 Firestore 原子快照。

## 查詢與索引

列表使用穩定排序與 cursor，不使用無界 offset 或全庫掃描。複合索引由實際查詢驅動並在目標環境驗證；避免為大型或不查詢欄位建立不必要索引。

## 生命週期

TTL 是延遲實體清理，不是授權或業務到期判斷。刪除父文件不會刪除子集合；清除流程需列出子集合、引用、唯一占用、衍生投影與操作收據。資料遷移採 expand、兼容讀寫、回填、切換與收縮，避免在 request path 做無界遷移。

## 備份與恢復

備份決策需定義 RPO、RTO、保存期限、權限與成本。還原後必須重新驗證授權、版本、唯一性與已處理操作，不能因資料可讀就直接恢復流量。

## 儲存庫共用資源編號

- 每個 Bot 管理的 Repository 共用一條資源編號序列，從 1 開始。例如 Issue #1、Discussion #2、Issue #3；不按資源類型另建計數器。另一個 Repository 從 1 開始。

- id 是不透明的隨機識別碼，number 是資源在空間內的可見編號。資料保存整數，畫面接入時才加 #。單一號碼能解析為資源類型與 ID，但不是授權憑證。各業務仍擁有獨立的狀態、權限與資料。

以下路徑均位於既有 channels/{channelId}/repositories/{repositoryId} 下：

| 路徑                            | 責任                                                                     |
| ------------------------------- | ------------------------------------------------------------------------ |
| counters/resources              | 唯一計數器，schemaVersion 與 lastNumber                                  |
| resourceNumbers/{number}        | schemaVersion、repositoryId、resourceType、resourceId、number、createdAt |
| resourceCreationReceipts/{hash} | schemaVersion、payloadHash 與原始編號對照結果                            |

- 資源類型使用穩定的小寫名稱；本基礎不接受客戶端直接選擇類型、操作者、ID 或號碼。收據鍵使用命名空間、可信操作者與防重鍵的 JSON tuple 做 SHA-256；channel 與 Repository 由保存路徑隔離。同一操作者在同一空間共用建立鍵範圍，跨類型重用鍵會回傳衝突。

- 初次成功寫入時才建立計數器，不由 Bot 建立 Repository 時預配。不回收、不重設、不因 Bot 離開或再加入而清零；不因刪除業務資料而刪除編號對照。上限為 JavaScript 可精確表示的最大整數 2^53−1。既有損壞計數器與下一號已存在時拒絕寫入，不覆蓋舊對照；不自動修復手動刪改的資料。

[編號交易](../../backend/internal/business/repository/adapter/firestore/numbering.go)

## 物件範本與套用收據

建立或改版的收據屬於範本所在的個人或本庫範圍，套用收據屬於目標儲存庫。各自以操作者及 `Idempotency-Key` 識別請求。相同鍵與正規化內容重送回原結果，內容不同回409 `idempotency_conflict`。套用收據包含原始節點 ID、名稱與父連結；結果回放代表建立時狀態，目前內容由既有 Object 讀取 API 取得。

Firestore adapter 在交易內重新檢查 Repository 的有效容器、範本快照及選填父節點，全部讀取先於寫入。Object 文件與套用收據一起提交；任何錯誤均不留下部分 Object 或成功收據。Repository 成員判定沿用目前服務的交易前存取查詢，未新增另一套授權機制。操作不呼叫 LINE，也沒有雲端部署副作用。

本庫範本位於 `channels/{channelId}/repositories/{repositoryId}`，個人範本位於同一 channel 的 `personalObjectTemplateLibraries/{ownerKey}`。`objectTemplates` 保存目前快照，個人範本改版時將上一版封存至該範本之下。`objectTemplateCreationReceipts` 與來源範本同庫，`objectTemplateInstanceReceipts` 和實際建立的 `objects` 位於目標儲存庫。未提供來源欄位的既有 Object 維持可讀，不搬移舊 `subobjects`。

錯誤包括400 `invalid_input`、404 `not_found`、409 `idempotency_conflict`／`version_conflict`，以及結果未確認時的503 `unavailable`。拒絕發生在寫入前或交易中；未知結果須保留原鍵與原內容重送，不能依掃描物件列表猜測成功與否。

[範本交易](../../backend/internal/business/object/adapter/firestore/templates.go)

## 群組出勤統計的讀取範圍

- 讀取既有 channels/{channel}/users 子集合資料，使用 DocumentRefs 包含只有子集合、沒有父文件的歷史使用者；不新增投影、資料回填、排程或環境變數。

- 每位使用者讀取本群 attendanceStates 與其活動班次；沒有活動班次時，以 clockOutAt 單欄範圍查詢今日已結束班次，於後端過濾本群。使用單欄自動索引，不新增複合索引；若正式環境已豁免該欄位索引，查詢會明確失敗，不能宣稱已驗收。

- 為限制首版讀取成本，單一channel最多列舉200個使用者，每人今日已結束班次查詢上限500筆（包含其他群組）；取得第201人或第501筆時整體拒絕，不回部分人數。看板讀取和本人檢查各有10秒期限，查詢失敗沿用既有Webhook錯誤／重送流程。此讀取方式適用小規模初版，非固定成本的大型群組聚合方案。

- 既有使用者的狀態與班次在同一唯讀交易讀取；使用者列舉在交易之前，不保證同時新增的使用者會納入該次看板。更新時間代表查詢完成時間，不是自動更新或持續即時訂閱。輸入指令才重新讀取。

[Firestore 看板查詢](../../backend/internal/business/attendance/adapter/firestore/board.go)

## 請假紀錄與歷史編號

- Firestore：channels/{channel}/leaveLaunches/{tokenHash} 保存群組與到期；channels/{channel}/users/{userHash}/leaveRequests/{requestKeyHash} 保存伺服器以 google/uuid 產生的 UUID v4 申請編號、群組、token雜湊、假別、UTC起訖、理由及建立時間。原始 token 不寫入假單文件；交易再次確認連結有效，重試不建立第二筆。

- 最近紀錄在本人子集合依 createdAt 排序，最多顯示 20 筆，沿用單欄索引，無須回填或新增環境變數。超出顯示範圍的登記仍保留。請假理由與去重紀錄的保存期限待定，首版不設定自動刪除（TTL）。

- 新申請的 API id 為 UUID v4；文件路徑保留 requestKeyHash 以維持原子防重複。舊資料沒有 id 時仍回傳原文件編號，不改寫歷史資料。UUID 產生失敗時交易不寫入。

[請假保存與查詢](../../backend/internal/business/attendance/adapter/firestore/leave.go)

## 專案歸屬與引用位置

- 四種引用共用 items 文件格式，將來源庫、來源類型、父物件及來源 ID 納入決定性鍵；不同庫的同名識別可獨立收錄。

| 位置                 | Project 路徑                                                                                                                                                                                      |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTTP                 | /api/v1/owners/{ownerId}/projects/{id}                                                                                                                                                            |
| Firestore 個人專案   | channels/{channelId}/users/{userId}/projects/{projectId}                                                                                                                                          |
| Firestore 組織專案   | channels/{channelId}/organizations/{organizationId}/projects/{projectId}                                                                                                                          |
| 舊群組專案相容資料   | channels/{channelId}/groups/{groupId}/projects/{projectId}；由既有 legacy adapter 隔離處理，不代表 canonical API 接受 Group Owner                                                                 |
| 儲存庫 LINE 群組綁定 | `channels/{channelId}/repositoryLineBindings/{digest(groupId)}`；保存 `groupId`、`repositoryId`、狀態及最後事件。它是 Group 與 Repository 關係的唯一權威來源，Repository 文件不重複保存 GroupID。 |

- Project 文件保存 ownerId；沒有初始來源時，對外回應省略 defaultRepositoryId。同一擁有者底下的 projectCreationScopes 保存建立收據，Project 底下的 linkedRepositories 保存庫連結；每筆引用以 sourceRepositoryId 定位來源。

[專案儲存實作](../../backend/internal/business/project/adapter/firestore/store.go)

## 標籤名稱與建立收據

- 資料位於既有 channels/{channelId}/repositories/{repositoryId} 下的 labels/{id}；labelNames/{hash} 保存名稱唯一綁定，labelCreationReceipts/{hash} 保存本空間、本人及建立鍵的收據。建立以同一交易保存標籤、名稱綁定與收據，所有讀取先於寫入。 Bot 停用保留資料、拒絕新讀寫；再次啟用沿用標籤。LINE 成員核驗是外部資格觀察，不宣稱與 Firestore 原子一致。錯誤只公開領域分類，SDK 原因保留在 adapter 私有值。

[標籤交易](../../backend/internal/business/repository/adapter/firestore/labels.go)
