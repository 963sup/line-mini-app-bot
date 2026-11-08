# 基礎出勤

首版處理群組定位上下班，以及 1:1 分群查詢本人紀錄。群組指令／Postback 建立短效請求，MINI App 取得位置並提交後才寫入出勤；部署與真實 LINE／Firestore 驗收需另外執行。
先完成 [基礎設施與接入](../development/foundation.md)，再開發本頁功能；
不要求在目前階段新增測試方案、案例或工具。

## 最小操作

| 操作 | 行為 |
| --- | --- |
| 上班 | 後端確認操作人員身分，保存上班時間，回傳操作結果。 |
| 下班 | 後端確認操作人員身分，記錄對應出勤的下班時間，回傳操作結果。 |
| 本人紀錄 | 僅回傳已驗證操作人員自己的出勤資料，至少包含上班時間與下班時間；尚未下班要明確呈現。 |

後端以可信身分識別操作人員，不接受前端自行指定他人的 user ID 作為授權。
打卡時間由後端產生，不直接信任前端傳入的時間。
資料保存失敗時不能回覆打卡成功。重複點擊或 LINE 重送不能悄悄新增重複紀錄。
只有上班沒有下班時，保留未完成狀態，不自行猜測下班時間。

## 不納入首版

排班、請假、加班審核、薪資、工作地點半徑判斷、人臉打卡、補登審核、管理員報表與通知推送。
不為上述功能預建模組或通用框架；既有 LINE 能力矩陣是參考，不代表全部都要實作。

## 已採用的首版規則

- Messaging API：群組輸入「打卡」取得 Flex 上下班按鈕，也可輸入「上班」「下班」。後端從已驗簽事件建立十分鐘有效的請求，綁定原操作者、群組與動作，回覆 MINI App 連結；此時尚未打卡。
- MINI App：使用者同意定位後，將 LINE ID token、requestId、latitude、longitude、accuracyMeters、capturedAt 送至 `POST /api/v1/attendance/punches`。後端驗證身分、請求歸屬與位置，再執行 Firestore 交易。拒絕定位不新增紀錄，不判斷半徑或在場真實性。
- 1:1 輸入「出勤紀錄」列出本人有紀錄的群組；輸入「出勤紀錄 編號」查該群最近五筆。群組查詢只回私訊引導，不開放 1:1 上下班。
- 各群組獨立計算；每人每群組同時只允許一段未完成出勤，一天可以多段。跨日下班關閉原本未完成紀錄。
- 顯示 Asia/Taipei，Firestore 保存 UTC。使用伺服器處理時間，不將事件時間當打卡時間；事件時間只用來拒絕早於最近處理事件的亂序操作。
- 重複上班回傳原上班時間；沒有未完成紀錄時下班明確拒絕。相同事件重送讀取原交易結果。
- 不需群組或人員核准名單，也不查好友資格。缺少 groupId、userId 或事件識別碼時不建立打卡請求。定位連結會出現在群組，其他成員開啟不能代原操作者打卡；結果顯示在 MINI App，不將精確座標推送群組。

## 啟用與資料

設定 `GOOGLE_CLOUD_PROJECT` 與 `LINE_MESSAGING_API_CHANNEL_ID` 後初始化出勤資料庫；任一未設定則出勤關閉，打卡回覆尚未啟用。
可選用 `FIRESTORE_DATABASE_ID`（省略為 `(default)`）。
沿用既有 Messaging API secret／access token 與 runtime ADC，執行帳戶需具有指定資料庫讀寫權限。
定位打卡另需 MINI App 的 Channel ID、LIFF ID、openid scope；Messaging API 與 MINI App 必須屬於同一 LINE Provider，否則兩邊 user ID 無法直接比對。LINE ID token 必須由伺服器驗證，不能相信前端自行指定的 userId/groupId。

LINE Developers Console 需開啟 Allow bot to join group chats，再邀請 Bot 入群。群組成員不必先各自加 Bot 好友；私訊查紀錄可先加好友開啟一對一。
舊有 `ATTENDANCE_MEMBERS_JSON` 設定不再讀取，不影響使用資格。私訊上下班只引導回群組，不建立另一套個人打卡。

資料位於 `channels/{channelID}/users/{SHA256(userID)}` 下：

- `attendance/{SHA256([groupID,eventID])}`：schemaVersion、groupID、clockInAt、clockOutAt（尚未下班則省略）。
- `attendanceStates/{SHA256(groupID)}`：recordID、lastEventTime。
- `attendanceEvents/{SHA256([groupID,eventID,action])}`：code、record（原操作結果快照）。

定位打卡的 schemaVersion 為 2；上班／下班各自保存 clockInPosition／clockOutPosition（經緯度、精度、採樣時間）。舊版紀錄仍可讀取。
`channels/{channelID}/punchRequests/{requestID}` 保存原 groupID、userKey（雜湊）、eventID、action、eventTime、expiresAt；相同事件重送不延長原請求期限。同一請求重試讀取原交易結果，不能新增第二次打卡。未完成請求逾時不接受新的寫入。

複合 ID 使用 JSON array 編碼後雜湊。事件結果、出勤紀錄與狀態在同一交易提交；回覆在交易外發送。
1:1 查詢固定本人路徑，以 groupID 等值條件取資料，在 Go 排序後只回最近五筆，不需新增複合索引。群組列表從本人紀錄取得，兼容舊資料；基礎雛形尚未加入分頁，因此讀取成本會隨紀錄數增加，正式擴量前需改為群組索引與有界查詢。
不儲存 reply token、完整事件或姓名；位置只隨打卡與去重結果保存，不建立持續定位歷史、不寫一般日誌。資料保存期尚待營運決策；去重結果不能先於可能重送期間刪除。punchRequests 可另設 expiresAt TTL；TTL 並非即時到期機制，程式仍自行檢查到期時間。

官方契約核對：LINE 文件來源 `line/line-developers-docs-source`，main revision `090500a12f9777995006d9fb5cafab8d83a792b2`；[LIFF 開啟與 query 傳遞](https://developers.line.biz/en/docs/liff/opening-liff-app/)、[伺服器驗證身分](https://developers.line.biz/en/docs/liff/using-user-profile/)、[同 Provider 的使用者 ID](https://developers.line.biz/en/docs/messaging-api/getting-user-ids/)。手機定位、Channel Provider 設定及真實 Firestore 交易仍需實機驗收。

LINE Developers Console 必須啟用 Webhook redelivery。寫入或出勤回覆失敗回傳 503，允許重送；
reply token 有效期與一次使用限制仍適用，去重保證不重複打卡，不保證每次回覆送達。遇回覆不明確時，私訊查詢本人紀錄。

## 後續產品變更

- 首期入口：Bot 或 MINI App 選定其一先接通；不預設兩者同步開發。
- 人員識別：LINE 身分如何對應允許打卡的人員，首次使用如何取得資格。
- 時間歸屬：顯示時區、跨日下班如何對應上班紀錄，以及是否允許一天多段出勤。
- 操作結果：重複上班／下班、沒有上班就下班時，明確拒絕或回傳既有結果的規則。

以上項目首版已依前節收斂；後續若改動，在本頁更新規則。
確認後才固定相應 API、Firestore 路徑／欄位、唯一性與原子更新方式；
不在基礎設施尚未穩定時一次建完全部功能。
