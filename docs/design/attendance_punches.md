# 打卡預檢、出勤月報與補正

## 打卡預檢結果

- Existing receipt; any action: replay original result despite changed state; no location.

- No open shift; clock in: needs_position.

- Working/resting; clock in: already_in and original start; no location.

- No open shift; clock out/break/confirm: not_in; no location.

- Working; clock out/start break/confirm breaks: needs_position.

- Working without break start; end break: not_resting; no location.

- Resting; start break/confirm: already_resting/break_incomplete; no location.

- Resting; clock out/end break: needs_position, subject to event ordering.

- Request predates last event; any unreceipted action: stale_event; no location.

- Expired unreceipted or non-owned request; any action: 403; no location/punch.

- Unknown backend state; any action: retryable error; no location.

- Preflight/submit share transactions and state machine; preserve code/message.accepted identifies current/original successful punches.needs_position reserves nothing, creates no event/receipt/shift and guarantees no later success.Persisted rejections replay forever for requestId; changed state needs a new group card.All successful actions, including breaks/confirmation, retain location evidence requirements.

[預檢入口](../../backend/internal/business/attendance/adapter/http/attendance_preflight.go) · [判定與收據](../../backend/internal/business/attendance/adapter/firestore/attendance.go)

## 定位失敗與送出後重試

- Location failure: no submit; retry same requestId through preflight first.

- Submit disconnect/timeout: unknown result; retry resolves receipt first.Display existing success without locating again.

- Leave/logout: cancel waiting and ignore late HTTP/location results by generation;already submitted transactions cannot be recalled.

- Same-page retries keep requestId and preflight.Navigation, session reset or full reopening clears page caches and revalidates launch/membership;a valid same card still exchanges the same requestId.

- After expiry, owners may replay existing receipts but unfinished requests cannot write.Location acquisition never extends deadlines.

## 本人查詢與月報範圍

伺服器以 Session Cookie 或 LINE ID token 驗證本人，不接受指定他人的 user ID。回應不快取；新版列表與詳情不公開精確位置或原始事件證據。

下表均為 GET，路徑前綴為 `/api/v1/attendance`。參數型別與回應結構見[出勤 OpenAPI](../../api/contract/attendance/records.yaml)。

| 用途         | 路徑               | 查詢參數                                             | 回應與範圍                                             |
| ------------ | ------------------ | ---------------------------------------------------- | ------------------------------------------------------ |
| 選群組       | `/groups/page`     | `pageSize`、`cursor`                                 | `items`（id、name）、`nextCursor`、`asOf`、`queryMode` |
| 查單群月摘要 | `/monthly/summary` | `group`、`month`（YYYY-MM）                          | 完整月份 Calculation、`count`、`asOf`                  |
| 列出本人班次 | `/records`         | `month`、`pageSize`、`cursor`；可選 `group` 篩選單群 | `items`、`nextCursor`、`asOf`、`queryMode`             |
| 查班次明細   | `/records/{id}`    | `group`                                              | 本人在該群組的班次詳情與版本                           |

- pageSize 預設20、範圍1–100；cursor 必須與本人、channel、群組／月份與排序版本相符，無效回400 invalid_cursor。班次排序為上班時間DESC、下班時間DESC（未下班的零值排在同上班時間的已下班班次之後）、文件ID DESC，以完整排序欄位消歧義，不使用offset。沒有下一頁時 nextCursor 為空。上一頁由瀏覽器保存已使用的cursor。

- 月摘要計算完整相關班次，不是目前一頁的加總。既有 /monthly、/groups 與 LINE 查詢保持契約，不能靜默截成一頁。即時分頁不提供跨HTTP請求快照保證；asOf只是產生時間。核准後重新查詢，資料變更時從第一頁重新載入。

- 月查詢包含與月份重疊的班次，計算時裁切為臺灣時間月初至下月初， 不把跨月整班時間全部算到上班月份。相同工作區間只累計一次並回報重疊異常。 月卡與 Web 顯示暫估合計（包含已確認工時）、出勤時數、休息待確認與無法計算筆數，正常工時／加班待認定。

## 補正請求與相容行為

- 新版Web在 POST /api/v1/attendance/corrections 帶一個 Idempotency-Key UUID。一次已確認的payload與key在結果未知時保持不變，重試不能產生新key；更改表單才建立新的提交。key不含使用者或敏感資料。

- 伺服器以channel、已驗證本人與key隔離回執，對固定版本的業務payload做正規化雜湊。時間以UTC正規化，breaks順序保留，server ID／CreatedAt不參與。申請與最小回執同交易建立；交易可能重跑，但不能建立第二筆申請。

- 空白、無效或重複header：400 invalid_idempotency_key。

- 無header：保留舊呼叫契約，但沒有跨請求去重保證；新Web一律帶header。

- 提交回應 status=accepted 只表示申請已接收，不宣稱仍待審；目前狀態使用既有狀態查詢取得。申請本人與審核權限仍由後端決定，禁止自己核准；核准保留 expectedRevision／expectedEventSequence 與不可變audit驗證。回執不保存token，未確定保存政策前不擅自設定TTL。

[補正 HTTP 處理](../../backend/internal/business/attendance/adapter/http/attendance_correction.go)

## 補卡核准與班次重疊

- 「補卡」處理漏上班、漏下班或整班皆漏；已有資料但時間不正確則使用「更正」。兩者是不同使用者功能與入口，但共用同一申請與審核核心。漏下班通常已有上班紀錄，仍屬補卡，不以「有沒有整筆紀錄」分類。沒有打卡資料不能自動認定缺勤或實際出勤。

- 員工送出時只建立申請；待審、駁回或撤回均不改有效工時。所有補卡及更正一律須人工核准。

- 完整角色管理延後；沿用伺服器管理的 attendanceReviewers 群組白名單，禁止自審。未配置審核者時維持待審，不自動核准。

- 核准同交易保存有效投影、before／proposal／after 稽核與決定；新補登班次無原始打卡事件或GPS，不偽造定位打卡。既有班次保留原始事件與版本檢查。

- 駁回保存理由、審核人及時間；撤回僅本人可操作。終態不能互換；相同決定重送不重複生效。撤回／駁回後要修改內容，另建申請。

- 核准時重查同員工、同群組的班次重疊；以既有 attendanceStates 群組文件協調正常打卡與核准交易。端點相接不算重疊，跨日／跨月不自動拒絕。

- 首版用 groupID 等值掃描檢查重疊，單員工／群組最多1000筆，超過時拒絕處理，不回傳部分驗證結果；不宣稱固定讀取成本。後續需依資料規模改為時段索引。

[核准交易](../../backend/internal/business/attendance/adapter/firestore/attendance_correction.go)
