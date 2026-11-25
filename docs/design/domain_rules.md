# 業務領域規則

## 用例設計概念

所有業務領域以相同概念檢查用例是否完整，但各 bounded context 仍擁有自己的模型、規則、授權與狀態轉移。這組概念不是新的領域、共用資料模型或通用互動框架，也不要求建立對應的共用型別、介面或服務。

| 概念          | 設計問題                         | 責任邊界                                                                  |
| ------------- | -------------------------------- | ------------------------------------------------------------------------- |
| Actor         | 誰正在提出操作或查詢？           | 行為主體由可信身分解析；Role、Membership 只可作為授權事實，不等同 Actor。 |
| Resource      | 操作或查詢的業務對象是什麼？     | 由所屬 bounded context 定義身分、生命週期與權威資料。                     |
| Action        | Actor 想執行什麼業務操作？       | 表達意圖；不等同操作成功後才成立的 Event。                                |
| Authorization | Actor 是否可以對 Resource 操作？ | 由後端依操作、資源與目前關係明確判定，不接受客戶端自行宣告。              |
| State         | Resource 目前是否允許這項操作？  | 描述業務狀態與合法轉移；不等同 Actor 的 Permission。                      |
| Event         | 哪個已提交的業務事實值得記錄？   | 只描述已成立的事實；唯讀查詢或無事件需求的操作不必為形式完整而產生事件。  |

狀態變更用例依下列順序分析：

```text
Actor + Resource + Action
          ↓
Authorization + Current State
          ↓
      Business Rules
          ↓
Resulting State + Event（需要時）
```

唯讀用例不修改資源，也不為符合模型而產生事件：

```text
Actor + Resource + Query
          ↓
     Authorization
          ↓
       Projection
```

用例設計至少應能回答 Actor、Resource、Action 或 Query、Authorization、相關 State、結果，以及適用時的 Event。此要求適用於 HTTP、LINE Messaging API、LINE MINI App、背景工作及其他 delivery channel。Delivery channel、Handler、Controller、LINE 群組或 MINI App 畫面只提供互動情境，不得取代資源歸屬、授權或狀態判斷。相同用例從不同入口觸發時共用業務規則，由各 Adapter 轉譯輸入與呈現結果。

下列概念不得混同：Actor 與 Role、Identity 與 Membership、Resource 與 Capability、Action 與 Event、State 與 Permission、Projection 與 Aggregate、External ID 與 Domain Identity、Authentication 與 Authorization。這些區分用於審查設計，不要求以一套跨領域基底型別實作。

## Attendance

| Term                            | Meaning here                                                                                                           |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Punch Event                     | Identified attempt with action, time/result. Rejected attempts also have results; only accepted actions change shifts. |
| Shift (worktime.Session)        | One person's clock-in/out within one group. Open shifts lack end times; this Session differs from login sessions.      |
| Interval                        | Start/end values for clipping reports, merging overlaps and calculating breaks; value-comparable without identity.     |
| Attendance Record               | Identifiable shift with times, event sequence and correction version; target of corrections.                           |
| Break confirmation              | Personal confirmation that breaks are complete. Empty lists distinguish confirmed no-break from unconfirmed.           |
| Confirmed net / Estimated total | Confirmed time requires complete verifiable evidence. Estimated total already includes confirmed time; never add them. |
| Correction                      | Reasoned changes to a specified record. Request, review decision and effective record are distinct.                    |
| Leave registration              | Personal group leave type, period/reason. Registration neither approves leave nor changes attendance.                  |

`GET /api/v1/attendance/daily-headcount` 以群組及 Asia/Taipei 工作日期查詢。後端核驗當下群組成員資格，只回傳人數與查詢時間。每位有一筆當天開始的班次的人計一次；跨夜班計在上班日。在規模上限或查詢失敗時回傳不可用，日誌顯示暫無資料。此人數與實際施工名單分開，查詢不寫入出勤或日誌。

### 工時試算與休息確認

- 09:00～18:00 的出勤為 540 分鐘。休息待確認時暫估 540 分鐘，即使已有休息區間也先保留整段。確認休息 60 分鐘後，已確認工時與暫估合計均為 480 分鐘，出勤仍為 540 分鐘。暫估僅供顯示，不得作為薪資結算依據。

- Go 核心以 ConfirmedNetMinutes 表達已確認淨工時，以 EstimatedTotalMinutes 表達含已確認部分的暫估總計；兩者不可相加。HTTP adapter 明確映射到既有 Minutes／EstimatedMinutes，不改 API 欄位、Firestore 文件或工時計算規則。

### 群組出勤統計與漏下班查詢

- 這是按指令查詢的最小版本，沒有背景排程、固定時間推播或全群催打卡。群內訊息對所有成員可見，因此看板不顯示姓名、座標或個別明細；本人檢查只回覆本次查詢者是否有未結束班次，不公開上班時間。

- 「現在下班」沿用原有短效連結、LINE本人／群組成員驗證、定位與交易寫入；開啟按鈕本身不會直接下班。記錄的是後端目前時間，不是推測的昨日下班時間。若是過去漏打，應從「補卡」入口或私訊「出勤紀錄」選擇原班次提出申請，不用現在下班來冒充過去時間。

- 連結與既有打卡卡片同樣允許原群組成員使用，但每人只能操作自己的班次；轉傳不會取得查詢者的班次操作權限。不是固定指向某一名員工的補登連結。

- 同日下班多次仍只計一人。下班後再次上班時，以目前活動班次判定上班中或休息中。

- 統計來自打卡紀錄，不是LINE即時成員名冊、排班、實際在場人數或缺勤統計。已離群者若仍有未結束班次，也可能出現在紀錄統計中；請先處理該班次。沒有打卡不能推論缺勤。「未結束」不自動等於確定漏打，第一版沒有班表或逾時判定。

### 請假與重送

- 請假可直接送出，無須取得定位。

- 假別為事假、病假、特休、其他。理由必填，最多500字；結束須晚於開始，單筆期間最多366天，開始年份限制2000至2100。允許登記過去時段，不計算應到日或可用假額。

- 第一版僅登記，不代表核准。 沒有審核、扣假、撤回、通知或自動發布群組結果；請另行聯絡主管確認。請假不改寫打卡、工時或更正資料，不以假單推定出勤。理由及假別不回傳群聊。

- GET /api/v1/attendance/leaves：登入本人最近20筆登記，按建立時間遞減，不代表完整歷史列表。

- 沿用既有 Cookie Session／Bearer 認證、CSRF 與 OpenAPI 請求驗證。後端以 LINE API 確認原群組成員，只有200才允許送出；非成員拒絕，API故障不寫入。

- 同一本人的 requestKey 在相同內容重試時回原登記；不同內容回409。前端在結果未知時鎖住草稿並重試原內容；成功後才可明確建立另一筆。

- 一般側欄入口供查閱本人紀錄；持群組連結才顯示登記表單。伺服器判定連結過期或失去成員資格後停止送件。回應逾時不代表失敗，先重新載入本人紀錄。頁面重開、連結過期後，應先確認最近紀錄再重新填寫；第一版不保存跨頁草稿，也不提供語意重複／時段重疊的自動拒絕。

[請假契約](../../api/contract/attendance/leave.yaml)

[打卡預檢、出勤月報與補正](attendance_punches.md) · [Attendance 實作](../../backend/internal/business/attendance/)

## Discussion

- 主題 title 去除首尾空白後 1–120 字元，body 最多 8000 字元；留言 body 去除首尾空白後 1–4000 字元。以 Unicode 字元計算。JSON 上限 65536
  bytes，不接受未知欄位；作者、ID、number、時間由伺服器取得。所有建立操作必須提供 Idempotency-Key，所有編輯操作必須提供 expectedVersion。

### 討論與留言分頁

- 討論依共用 number 遞減，一頁掃描二十筆候選後依標題搜尋；允許空頁續查。留言依 createdAt 及 ID 遞增分頁，每頁二十筆；新增留言不改主題內容版本。兩者皆有範圍綁定游標，
  前端不以單頁資料推算總數。

[Discussion 實作](../../backend/internal/business/discussion/)

## Enterprise

| 資料或操作                   | 規則與驗收條件                                                                                                                                                      |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enterprises/{enterpriseId}` | 企業集中管理旗下組織與 Enterprise Team。Organization 可獨立存在，加入企業後以 enterpriseId 指向一個 Enterprise。                                                    |
| 指派企業 Team                | 具企業 Team 指派權的管理者指定旗下 Organization。後端核對 Team 的管理企業與目標組織的企業相同，再依組織與 Team 的組合鍵寫入；跨企業一律拒絕，重送不產生第二份名單。 |
| 移除指派                     | 刪除指定組織的指派。依賴這筆指派的授權即失效，其他組織的指派與企業成員名單維持有效。                                                                                |

[Enterprise 實作](../../backend/internal/business/enterprise/)

## Issue

Repository 提供資源範圍，Issue 維護自己的開啟與關閉狀態。問題建立者可以直接關閉開啟中的問題，也可以重新開啟；Go 核驗操作權與版本。專案內的操作與驗收見[專案項目進度](#project)。

相容實作沿用 API 的 `completed` 表示已關閉，保留既有 `in_progress`
與變更紀錄。新版入口直接關閉或重開，舊客戶端仍可送出原有轉換。

### 問題列表與建立收據

- 一次最多掃描二十筆候選，搜尋以正規化標題、不區分大小寫的子字串比對，再套用狀態篩選。可能回傳空頁但仍有 nextCursor，前端必須提供繼續載入；不將單頁筆數作總數，不新增複合索引。
  游標無效或換空間／身分／條件後沿用會被拒絕。 建立鍵與同空間同操作者的其他資源共用範圍；相同鍵及內容回同一 ID／number 的目前快照，內容不同回 409。建立雜湊使用
  adapter 明確編碼的 v1 正規化 title／body，不雜湊 HTTP
  DTO。

[Issue 實作](../../backend/internal/business/issue/)

## Journal

新版已支援從儲存庫建立與查閱。首版以啟用儲存庫既有存取判定核驗操作者：個人庫限本人，群組庫限即時核驗的成員。保存前交易會再檢查庫狀態及擁有者。舊群組日誌保留歷史查閱。建立 API 只接受 `repositoryId`，傳入 `groupId` 或缺少主庫會被拒絕。Feed 與 Dashboard 可切換為儲存庫和工作日期。日誌可複製僅含入口的連結，直接發送群組摘要與正式渠道驗收仍待接線。撰寫時可另選群組查詢[出勤參考](#attendance)。

[Journal 實作](../../backend/internal/business/journal/)

## Membership

組織頁預設開啟成員分頁，另有團隊與管理分頁。首版成員名單只對當前成員開放，使用組織範圍內的成員代碼供辨識。使用者可分享自己的代碼供 Owner 核對；原始 LINE 使用者 ID 不回傳。一般成員可退出，Owner 可移除他人，現任 Owner 必須先移轉才能退出。

現行 HTTP 契約為 `GET /api/v1/organizations/{organizationId}/members`、`DELETE /api/v1/organizations/{organizationId}/membership`（本人退出）及 `DELETE /api/v1/organizations/{organizationId}/members/{memberId}`（Owner 移除）。兩個異動都要求 `Idempotency-Key`。交易中以同一組織文件重新判斷權限，原子更新 `memberUserIds`、記錄最近退出時間並保存稽核與重送收據。最近退出時間阻擋之前簽發、尚未兌換的邀請重新開權。退出後重載組織選單，移除後重載名單。

[Membership 實作](../../backend/internal/business/membership/)

## Object

父物件與子項都是同一個儲存庫內的 `Object`。子項透過 `parentId` 指向父物件，兩者分別維護數量、單位、完成數與版本。

- 使用 `GET /api/v1/repositories/{repositoryId}/objects?parentId={id}` 分頁列出直接子項。
- 建立子項使用 `POST /api/v1/repositories/{repositoryId}/objects`，在請求本文帶入 `parentId`。編輯子項使用其自身的 Object ID 與版本。
- 專案與工作日誌都引用子項自己的 Object ID。

ObjectTemplate 是 Object 領域內的獨立聚合。本庫範本歸屬單一儲存庫，個人範本歸屬持有人。套用範本時，應用層依確認的節點建立獨立的 Object 樹。

[物件範本版本與套用](object_templates.md) · [Object 實作](../../backend/internal/business/object/)

## Organization

| 資料或操作                       | 規則與驗收條件                                                                                                                                                                                                   |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `organizations/{organizationId}` | 建立時由後端產生 UUID。每個組織恰有一位現任 Owner，唯一判定欄位為 `ownerUserId`。`name` 可修改，`createdBy` 保存建立者。                                                                                         |
| 建立組織                         | 已登入使用者建立 Organization，後端從已驗證身分填入 ownerUserId 與 createdBy。同一位使用者可建立多個組織。                                                                                                       |
| 移轉所有權                       | 只有現任 Owner 能指定有效的現有成員為新 Owner。同一交易重新核對現任 Owner 與接任者的成員資格，更新 `ownerUserId`、Owner 版本，並保留舊 Owner 為一般成員。舊 Owner 失去管理權，`createdBy` 與 Team 歸屬維持原值。 |

[Organization 實作](../../backend/internal/business/organization/)

## Project

每個專案恰有一個 Owner，類型只能是使用者或組織。LINE 群組可提供入口，但群組身分本身不具備專案所有權。一個專案可追蹤零至多個儲存庫的工作。建立時可以先建空白專案；若從儲存庫進入，可預選有權使用且屬於同一擁有者的初始來源。之後可加入其他有權讀取的來源。各項引用沿用來源庫的讀取權限。

每個 Project item 各自保存待辦、進行中或完成。缺少欄位的既有項目讀為待辦。來源狀態依[問題生命週期](#issue)、工程物件或巡檢規則讀取。畫面分別標示「本專案進度」與「來源狀態」。更新進度沿用 Project 編輯權、版本檢查與交易；來源讀權仍由來源庫核驗。

### 專案檢視與跨 Owner 複製

專案建立者可在詳情頁編輯自訂欄位、選項及檢視，與專案內容一併儲存。每個檢視依專案內的穩定欄位 ID 建立關聯，重新命名欄位後仍指向同一欄位。移除欄位前須先改選或移除引用它的檢視。後端在同一筆專案版本更新中驗證設定；版本衝突時須重新讀取。一般讀者可以查看設定。

使用者可選擇另一個有權建立專案的 Owner 作為目的地。後端須同時核驗來源專案讀取權與目的地建立權。成功後目的地清單出現具有新 ID 的獨立副本；原專案保留原 ID、Owner 與內容。副本的存取權由目的地決定，引用的儲存庫資料仍須逐庫核驗。

複製以來源專案作為設定範本。使用者選擇自己的帳戶或有建立權的組織，並指定新專案名稱。副本保留檢視、自訂欄位與選項；檢視參照的欄位須對應到副本自己的欄位。原有工作項目、外部引用、巡檢紀錄、協作者、團隊與儲存庫連結均不複製，副本的初始來源為空。

### 工程驗收送件

GET／POST /api/v1/owners/{ownerId}/projects/{id}/references/{itemId}/inspection-submissions 分別讀歷史與提交。資料保存在引用下的 inspectionSubmissions。施工進度維持原值，有送件紀錄的引用不可移除。

專案引用回傳 `inspectionMode`。工程物件與子物件可查閱驗收紀錄，來源可用時可在空紀錄下提交；問題與討論不提供驗收入口。實際讀寫時，後端仍會核對儲存庫存取、專案連結及既有驗收紀錄。

[送件契約](../../api/contract/project/inspection_submissions.yaml)

[Project 實作](../../backend/internal/business/project/)

## Repository

成員可讀取所屬儲存庫的事項狀態、討論、里程碑與相容專案數。此數量只計入以該庫為初始來源（`defaultRepositoryId`）的專案；後來才連結此庫的跨庫專案不計入。Project 由個人或組織擁有，這項統計不表示 Repository 擁有 Project。

Repository Owner 只能是 User 或 Organization。LINE Group 是可選的互動綁定，不是 Owner；群組成員資格不直接等於 Repository 權限。首次有效群組事件以已驗證 Actor 當下明確選定且仍具資格的 Organization 作為 Owner，沒有有效組織選擇時才由 Actor 個人擁有，並建立獨立 Group Binding。Bot 離群只停用 Binding，Repository 本身維持啟用。Organization-owned Repository 依組織成員資格授權。資料直接採新欄位與新 Binding 集合，不讀取舊 `OwnerID=GroupID` 或 `repositoryOwnerBindings` 格式。

### 地址

地址是 Repository 擁有的可選 Value Object，由 ISO 兩碼國家代碼、郵遞區號及地址行組成。成員僅能修改當下有權存取且仍啟用的儲存庫；寫入須帶目前版本，Firestore 交易再次核對儲存庫身分、擁有者、狀態與版本，過期寫入回 409 且不改變資料。OpenAPI 是傳輸契約，Go 是驗證與狀態轉移的權威，前端只使用產生的 API client 編輯及呈現。

### Wiki 頁面與修訂

首版提供目錄、建立、閱讀、修改及版本歷史。每頁有固定 page ID、庫內唯一標題、Markdown 內容及目前版本。首頁可由目錄進入；頁面互引使用固定 ID，改名後舊連結仍可開啟。每次有效修改由 Go 在同一交易更新頁面與修訂，保存作者、時間、標題和內容。提交時核對預期版本，過期回 409 並保留編輯草稿供使用者比對；建立請求重送沿用同一識別鍵。首版以文字與連結為主。

成員甲建立 SOP，成員乙修改標題與步驟後，原連結可開啟新版，歷史仍可查看舊版及修訂作者。兩人從同版提交，僅第一筆成功。離群者持有舊連結也無法取得正文或歷史；相同標題可分別存在不同 Repository。由 OpenAPI 定義正式操作與資料格式，Go 負責存取與版本規則，React 負責閱讀及編輯。

目錄與歷史每頁最多 20 筆，歷史依版本由新到舊。目錄只傳標題與修訂資訊，開啟頁面才取得正文。首版標記支援標題、段落與 HTTP(S) 連結。操作格式見 [Wiki OpenAPI](../../api/contract/repository/wiki.yaml)。

### 洞察統計的讀取範圍

| 項目                    | 規則                                                                |
| ----------------------- | ------------------------------------------------------------------- |
| Actor／Resource／Action | Session／Bearer 核驗本人，讀取所屬儲存庫的完整統計                  |
| Context／Authority      | 啟用儲存庫與當下群組成員資格；只可讀取有權範圍                      |
| Policy／Decision        | 統計完整集合，不以單頁筆數作總數；沿用各功能自己的狀態              |
| Result                  | 真實數量與觀察時間；讀取失敗顯示錯誤，不偽裝為 0 或沿用過期成功數字 |
| 併發／一致性            | 即時查詢，無快取或背景工作；不提供跨請求統計快照保證                |
| Event／Audit            | 唯讀，不建立額外寫入或外部通知                                      |

[統計查詢](../../backend/internal/business/repository/adapter/firestore/statistics.go)

[Repository 實作](../../backend/internal/business/repository/)

## Summary

- summaryJobs/{hash(eventID)}：Job 欄位。Sources 僅存訊息 ID，不複製原文；Result 僅暫存以供 Push 重試，工作結束清空 Result。

- 訊息以穩定 message ID 原子建立；命令以 event ID 原子建立工作；Cloud Tasks 也使用固定 task ID。佇列建立失敗回傳503，由 LINE 重送恢復。

[Summary 實作](../../backend/internal/business/summary/)

## Team

Enterprise Team 指派只是組織的候選來源；在組織列表中仍須先具備該組織可見資格。LINE 群組資料與原有群組操作繼續適用自身的資格規則。

Team 管理歸屬操作拒絕沿用獨立模式團隊的舊管理者欄位。維護成員資格與變更團隊管理歸屬是兩項不同能力。

[Team 實作](../../backend/internal/business/team/)

## Weather

- Bot 在群組收到 LINE 內建「分享位置」訊息後，直接回覆目前有效時段的縣市天氣預報。 私訊位置、地圖網址、一般文字不觸發。沒有名單、訂閱、排程或位置儲存。

- 使用 F-D0047-091 縣市未來一週預報，讀取目前有效的「天氣預報綜合描述」，直接呈現官方溫度、降雨及天氣文字。 先用 LINE 位置地址中的縣市名稱比對（兼容台／臺），代表點距離不得超過 100 公里。 沒有相符地址時，以球面距離選最近的官方代表點，最多 60 公里，回覆明確標示「鄰近縣市預報」。 這是基礎區域近似，不是行政區邊界判定、鄉鎮定位或座標即時觀測；山區與縣市邊界尤其需注意。 不支援的遠端位置、過期預報或上游錯誤不回傳虛構天氣。

[Weather 實作](../../backend/internal/business/weather/)
