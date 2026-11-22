# 組織與工作區選擇

## Organization（組織）

本頁保存 Organization 的資源與導覽規則；邀請、名單、退出與移除的操作契約統一維護於 [Membership 的組織入口](../../proposals/membership.md#首版-bot-私聊入口)。現行 legacy 流程與未來獨立 Membership 模式須分別驗收。

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `organizations/{organizationId}` | 建立時由後端產生 UUID。每個組織恰有一位現任 Owner，唯一判定欄位為 `ownerUserId`。`name` 可修改，`createdBy` 保存建立者。 |
| 建立組織 | 已登入使用者建立 Organization，後端從已驗證身分填入 ownerUserId 與 createdBy。同一位使用者可建立多個組織。 |
| 移轉所有權 | 只有現任 Owner 能指定有效的現有成員為新 Owner。同一交易重新核對現任 Owner 與接任者的成員資格，更新 `ownerUserId`、Owner 版本，並保留舊 Owner 為一般成員。舊 Owner 失去管理權，`createdBy` 與 Team 歸屬維持原值。 |

## 側欄切換組織（目標規格）

採用一個 shadcn 選單。企業是分組標題，組織是可選的工作區。桌面與手機共用下列表現。

| 畫面狀態 | 側欄呈現與操作 | 資料來源 |
|---|---|---|
| 身分驗證中 | 顯示不可操作的「正在驗證身分…」。確認身分後才載入組織。 | 共用 session 狀態 |
| 收合 | SidebarMenuButton 第一行顯示目前組織（例：台北營運），第二行顯示企業（例：台灣總公司）。可存取名稱為「切換組織」。 | selected.name 與 selected.groupLabel |
| 展開 | DropdownMenu 以「切換組織」為標題。每個分組顯示企業名稱，下面列出組織單選項目；目前組織打勾。點企業名稱開啟企業頁，點組織項目切換組織。 | 按 groups 與 organizations 的回傳順序繪製，用 organizationId 比對選取狀態。 |
| 獨立組織 | 無所屬企業的組織列在「獨立組織」分組，收合時第二行也使用此名稱。 | 後端提供 enterpriseId 為 null 的單一分組。 |
| 同名組織 | 跨企業以分組區分；同分組內重名時，在項目後顯示後端提供的辨識文字。 | option.detail，例如可區分兩者的 ID 片段。 |
| 零或一個組織 | 零個顯示「尚無可用組織」；只有一個時保留組織與企業兩行，省略展開操作。 | selected 與 canSwitch |

### API 與後端型別

以下操作，沿用登入 Session 與共用 APIError。[OpenAPI 契約與 Go／TypeScript 生成流程](../../../api/README.md)為實作依據。

| 操作 | 請求 | 成功結果 |
|---|---|---|
| GET /api/v1/me/organization-context | 由後端從登入憑證取得 userId | 200 OrganizationContext，列出本人擁有或具有效成員資格的組織。先採用仍有效的上次選擇，否則取第一組的第一個組織。後端依企業名稱、企業 ID 排序，獨立組織組排最後；組內依名稱、組織 ID 排序。僅回傳含可見組織的分組。 |
| PUT /api/v1/me/organization-context | SelectOrganizationRequest | 後端重新核對資格，依 userId 保存選擇並回 200 OrganizationContext。重選同一組織結果一致。未登入回 401；格式錯誤回 400；目標不存在或無權使用統一回 404，保留原選擇。 |

| OpenAPI 型別 | JSON 欄位與對應 Go HTTP DTO 型別 | 契約 |
|---|---|---|
| OrganizationContext | groups → []OrganizationGroup；selected → *SelectedOrganization；canSwitch → bool | groups 永遠為陣列。無可用組織時回 []、null、false。至少兩個可用組織時 canSwitch 才為 true。 |
| OrganizationGroup | enterpriseId → *string；label → string；organizations → []OrganizationOption | enterpriseId 為企業識別碼或 null；label 為可直接顯示的企業名稱或「獨立組織」。 |
| OrganizationOption | organizationId → string；name → string；detail → *string | organizationId 為 UUID，各組間唯一；detail 通常為 null，同組同名時必須提供。 |
| SelectedOrganization | organizationId → string；name → string；groupLabel → string | 必須對應 groups 中的一個選項；name 與 groupLabel 由後端從同一份清單投影，側欄可直接顯示。 |
| SelectOrganizationRequest | organizationId → string | PUT 唯一輸入為目標組織 UUID。企業歸屬與使用者身分由後端解析。 |

以上 JSON 鍵皆必填；Go 指標欄位允許 JSON null。前端使用 OpenAPI 生成的 TypeScript 型別，GET 載入與 PUT 切換成功時都用整份 OrganizationContext 更新畫面。後端一次提供分組、選取摘要與可切換狀態，前端只需一個請求即可呈現側欄。

後端用例負責清單、排序、預設值、授權與偏好保存；HTTP adapter 轉成生成的 DTO。後續組織資源請求須明確帶目標 organizationId，後端逐次核對資源歸屬與操作權限。切換選單只更新個人偏好。

Organization 查詢只組合可見組織、選擇偏好與 Enterprise 提供的導覽事實。Bootstrap 注入 Enterprise 的資料與開通能力；兩者仍參與呼叫端的同一 Firestore 交易。Organization、Team 與 Enterprise 對現行組織成員欄位的授權只適用於啟用的 legacy 組織。獨立 Membership 模式的正向讀寫須在切換前接妥。

組織頁的「設定」提供 Owner 移轉。畫面從 People 名單選擇接任者，確認後以 `memberId` 呼叫 `PUT /api/v1/organizations/{organizationId}/owner`，成功回 204 並重讀名單與側欄。後端在指定組織的交易內解析成員代碼；既有使用 `userId` 的呼叫仍回 200。兩個目標欄位只能提供其一。

組織頁另以 API `GET /api/v1/organizations/{organizationId}/teams` 讀取清單。後端先驗證組織存取資格，再合併本人可見的本地 Team 與有效企業指派，依 teamId 去重。成功回傳 `OrganizationTeamsResponse { teams: OrganizationTeamOption[] }`；每個選項含 `teamId`（LINE groupId 字串）、`name`、`managementScope`（organization 或 enterprise）。無可見 Team 時回空陣列。Go DTO 與前端型別同樣從 OpenAPI 生成。切換側欄只需帶入新 organizationId 重新查詢，前端不合併清單，也不從 Team 指派推算可切換的組織。

### 前端責任與驗收

- 側欄呼叫共用 API client，直接使用生成的 TypeScript 型別與回傳狀態。兩個側欄入口共用同一切換元件。前端只保留載入、送出中與錯誤狀態，不另存選取偏好或推算權限。
- 首次載入時顯示載入狀態；空清單顯示「尚無可用組織」。單一組織直接顯示名稱，多個組織可開啟選單。重新載入後以 GET 的結果恢復選取。直接開啟組織連結或按上一頁時，由 URL 決定要查看的組織頁，保留伺服器上的組織偏好。
- 展開選單只顯示選項。選取另一個組織時送出一次 PUT，儲存期間暫停重複送出。成功後沿用回傳的 context，手機收合側欄，再以 Next.js 客戶端導覽開啟組織頁並載入新團隊資料。導覽沿用既有 session；導覽失敗時可重試開頁。
- 儲存失敗時顯示錯誤並重新取得後端狀態；若成員資格已撤銷，依最新清單更新畫面。驗收應確認單純展開不送 PUT，成功切換只送一次 PUT、無額外 POST /api/v1/session，且側欄直接顯示新組織名稱。其他驗收涵蓋空清單、同名組織、鍵盤與手機切換、重新整理、失去資格與舊回應晚到。



## 獨立資格的交付邊界

未來設計：建立新組織時，Organization、範圍狀態與首位 Owner 關係 MUST 原子建立。既有組織依 [Membership 切換契約](../../proposals/membership.md#儲存與遷移方案)準備關係並切換；不得只改模式欄位。切換前 MUST 接妥 Bot 邀請、People、Owner 移轉、退出／移除、組織選單及 Team／Enterprise 消費端的正向讀寫。切換後舊欄位不能補授權。

首版沿用 Bot 和 MINI App 的現有入口。新組織開通獨立模式時，首位 Owner 取得有效關係和初始資格世代。已驗證使用者接受有效邀請後，由同一交易建立或恢復成員關係。退出或移除由原 People 入口執行，交易內重驗權限並結束關係。Owner 移轉只更新權威指標，原 Owner 仍是有效一般成員。重送、版本和世代沿用 Membership 共通契約。

Team 透過交易內的資格讀取契約取得目前 Owner 判定、首任候選人的有效關係和資格世代。MINI App 從服務端提供的候選名單選人，請求使用組織成員代碼。Team 不寫入 Organization 關係，也不把現行 legacy 欄位當作獨立資格來源。驗收由管理者建立或切換一個測試組織，完成邀請、退出及重新加入後，再驗證 [Team 開通授權](../team/design.md#群組開通授權)。

驗收須從正常入口完成建立、邀請、接受、名單、移轉及退出／移除，並核對重送不重新開權、移轉與退出併發仍有有效 Owner、撤權後選單及下層資格立即失效。切換後重新加入產生新世代，不能恢復舊 Team 資格。資料準備、切換執行入口及回復工具尚待實作與環境驗收；未完成前 MUST 維持 legacy 模式。

[跨領域歸屬契約](../../contracts/organization-team-ownership.md) · [領域入口](README.md)
