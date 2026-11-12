# 群組 Repository

## 範圍與資料

每個 LINE 群組一個 Bot 管理的空間，具有獨立 `id` 與唯一 `groupId`。
不建立人類擁有者、角色、手動建立、換綁或刪除能力。Bot 管理容器不授予使用者審核或讀取他人出勤的權限。

資料位於 `channels/{channelId}/repositories/{id}`，保存 id、groupId、name、status、version、createdAt、updatedAt；內部另存最後事件時間及 ID。
`repositoryGroupBindings/{groupId}` 與容器在同一交易建立；ID 隨機產生，重送與併發沿用既有群組綁定。
ACTIVE / INACTIVE 表示 Bot 群組入口狀態。離開保留資料，再加入沿用 ID。舊事件不得覆寫新狀態；同時間停用優先。版本只在公開狀態或名稱改變時遞增。
名稱是加入或首次互動時取得的群組名稱快照，查詢失敗時允許空白，後續事件會重試補齊；UI 顯示「群組空間」。不宣稱即時同步群組更名。

## 操作契約

| 欄位 | 觀察群組事件、建立／更新 |
| --- | --- |
| Actor / Resource / Action | 驗證簽章的 Bot webhook；群組對應的 Repository；Observe |
| Context / Authority | Messaging API channel 範圍、事件時間與 ID；只接受 active 模式群組加入、離開、訊息、postback、成員加入事件 |
| Policy / Decision | groupId 格式有效；交易內讀取唯一綁定與現有容器，避免重複及舊事件覆寫；不消耗原事件，後續摘要與出勤照常執行 |
| Result | 原子建立／更新容器與使用者候選索引；不搬遷或改寫既有摘要、出勤資料 |
| 重送／併發 | 唯一綁定交易、固定新 ID 於重試外產生；相同事件不增加版本 |
| Event / Audit | 保留最後來源事件 ID、來源時間及公開更新時間；不保存 token、訊息內容或個人 profile |

| 欄位 | 列表／詳情查詢 |
| --- | --- |
| Actor / Resource / Action | 現有 Cookie Session 或 Bearer 驗證的使用者；Repository；List / Get |
| Context / Authority | 候選來自可信群組互動，不代表成員資格；每次以 LINE group-member-profile 核驗目前成員 |
| Policy / Decision | 只顯示 ACTIVE 且核驗成功的容器；不存在、停用或無權詳情統一 404；LINE 非 200/404 或無效 profile 為 503，不能推定有權 |
| Result | 不可快取的基本資料；沒有摘要內容、他人出勤或管理能力 |
| 重送／併發 | 讀取無副作用；成員關係為查詢當下的外部觀察，不宣稱跨 LINE/Firestore 原子快照 |
| Event / Audit | 不為一般讀取另存個人資料或事件 |

## API 與前端

- `GET /api/v1/repositories?cursor=...`：一次讀取本人最多十筆互動候選，再核驗目前資格；可能回傳空頁及 nextCursor，需繼續下一頁。
- `GET /api/v1/repositories/{id}`：即時核驗群組成員；格式錯誤／不存在／無權均不揭露資料。
- API 契約：[Repository](../../api/contract/repository/repositories.yaml)，由 [OpenAPI 入口](../../api/contract/openapi.yaml) 組合；生成 Go/TS DTO，不手改產生碼。
- UI：側欄「空間」可展開已驗證使用者可開啟的空間，選取後在 `/miniapp/repositories?selected=...` 顯示既有詳情。原頁不再重複顯示另一份列表；既有 `?id=...` 及 `repository`／`liff.state` 群組入口仍相容。
- `business/repository/RepositoryContext.tsx` 共用列表、分頁與詳情載入：一般功能頁首次展開側欄才讀列表，空間頁自動載入；列表回空但有 nextCursor 時保留「載入更多空間」。錯誤清空舊項目並提供重試，不將候選或停用空間當作可開啟。
- 側欄一般選取使用乾淨 `selected` 參數，不沿用打卡 launch／其他表單 query；既有 `id`／`repository` 群組專用入口只讀指定詳情，不讀一般列表、不於失敗後退回其他群組。這些參數只在空間頁解析，不把其他功能的 `id` 當空間 ID。
- 列表與詳情各自取消舊請求，並以 Session epoch、本人及作用域隔離可見狀態；登出／重新驗證立即隱藏舊資料，即使 transport 忽略取消也不得跨身分顯示。點選空間仍重新呼叫詳情 API 核驗，不以列表快取授權。
- 手機選取空間後關閉抽屜；預覽與定位頁保留正式空間頁連結，不建立虛假的本人空間資料。無 JavaScript 時提供原路由連結及啟用提示。
- 出勤連結只帶既有 group 查詢條件，不新增授權；摘要仍回原群組操作。
- 從成員加入、本人群組訊息或 postback 建立列表候選。未互動的既有群組成員需先互動；這不是完整群組成員清單。

## 啟用與遷移限制

沿用 GOOGLE_CLOUD_PROJECT、FIRESTORE_DATABASE_ID、Messaging API 設定；缺少設定時查詢回 503。
MINI App 與 Messaging API 使用者識別須在相同 LINE provider 下對應；無法匹配時不能通過成員核驗。
既有群組收到下一筆支援的可信事件時建立，不執行正式資料掃描或批次遷移。Summary / Attendance 歷史資料仍以 groupId 保存；本階段只建立容器、解析介面與讀取 UI，後續領域寫入接入另行實作。

## 查證與驗收

- LINE 官方來源 main：`dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54`，查閱 group chats、Messaging API reference 的 join/leave/group-member-profile，確認 404 不授權及成員 profile 的 userId。
- Context7 `/googleapis/google-cloud-go`：Firestore RunTransaction 重試及先讀後寫；對照專案 firestore v1.29.2。Google Developer Knowledge 未在此聊天提供工具，使用官方 Go client 文件；未做雲端部署或資料操作。
- Context7 `/vercel/next.js/v16.4.5` 與安裝版 static-export 指引：沿用 Pages Router、客戶端載入及靜態輸出。
- Graphify 已定位 webhook、server、identity 及既有群組查詢；Repomix 使用本次限定範圍產物 `/tmp/repository-before.xml`，未涵蓋的指引與文件直接閱讀。
- 本機檢查結果於交付訊息列出；LINE 真實群組事件、正式 Firestore 併發與 iOS/Android 登入需另行環境驗收，不能以建置通過替代。

本次本機驗證：前端靜態建置與型別檢查、12 個既有前端測試、OpenAPI 產生碼同步／相容性、設定檢查、Go vet／既有 race 測試／build／lint／格式與漏洞檢查通過。獨立暫存 smoke 使用受控 store／身分與 LINE transport，確認未登入、非成員、停用、錯誤 cursor、profile 不符與 404／429 的拒絕語義；不代表 Firestore 真實交易或 LINE 整合驗收。未新增正式測試案例或工具。

LINE「選單 → 請選擇功能」與功能說明提供「空間」URI 入口，以既有 LIFF URL 加 `repositories=1`，首頁解析一般 query 或 liff.state 後導向空間頁。

群組空間入口由可信 groupId 查得 Repository ID，連結帶 `repositories=1&repository=<id>`。一般 query 與 liff.state 均保留目標 ID；專用入口只查詳情，重新載入仍查同一 ID，失敗不退回列表。私訊入口維持本人列表。後端仍依目標容器 groupId 即時核驗成員；連結不是授權憑證，修改 ID 不授予其他群組權限。

群組專用入口本機驗證：以實際 Go 靜態頁路由、受控 LINE SDK／API 回應檢查一般 query、liff.state 與 404 情境；三者均只請求指定 Repository 詳情，重新載入不切換到列表，390px 無水平溢出。這不取代 LINE 裝置驗收。

## 後續模型設計

內部共用編號基礎見 [Repository 共用資源編號](repository-numbering.md)。目前沒有接入業務建立入口，不改變本頁 Bot 管理、公開查詢或成員核驗規則。

Repository 的實體、查詢、集合與 Connection 分頁命名，以及 Issue／Discussion 的歸屬規劃，見 [業務模型設計](../architecture/domain-model.md)。該頁為後續開發參考，不變更本頁既有 REST、成員核驗與 Bot 管理模式。
