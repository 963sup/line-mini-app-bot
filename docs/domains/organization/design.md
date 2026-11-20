# 組織與工作區選擇

## Organization（組織）

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `organizations/{organizationId}` | 建立時由後端產生 UUID。每個組織恰有一位現任 Owner，唯一判定欄位為 `ownerUserId`。`name` 可修改，`createdBy` 保存建立者。 |
| 建立組織 | 已登入使用者建立 Organization，後端從已驗證身分填入 ownerUserId 與 createdBy。同一位使用者可建立多個組織。 |
| 移轉所有權 | 只有現任 Owner 能指定有效的現有成員為新 Owner。交易內重新核對現任 Owner 並更新 ownerUserId；舊 Owner 隨即失去 Owner 權限，createdBy 與 Team 歸屬維持原值。 |

## 側欄切換組織（目標規格）

採用一個 shadcn 選單。企業是分組標題，組織是可選的工作區。桌面與手機共用下列表現。

| 畫面狀態 | 側欄呈現與操作 | 資料來源 |
|---|---|---|
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

組織頁另以 API `GET /api/v1/organizations/{organizationId}/teams` 讀取清單。後端先驗證組織存取資格，再合併本人可見的本地 Team 與有效企業指派，依 teamId 去重。成功回傳 `OrganizationTeamsResponse { teams: OrganizationTeamOption[] }`；每個選項含 `teamId`（LINE groupId 字串）、`name`、`managementScope`（organization 或 enterprise）。無可見 Team 時回空陣列。Go DTO 與前端型別同樣從 OpenAPI 生成。切換側欄只需帶入新 organizationId 重新查詢，前端不合併清單，也不從 Team 指派推算可切換的組織。

### 前端責任與驗收

- 側欄呼叫共用 API client，直接使用生成的 TypeScript 型別與回傳狀態。兩個側欄入口共用同一切換元件。前端只保留載入、送出中與錯誤狀態，不另存選取偏好或推算權限。
- 首次載入時顯示載入狀態；空清單顯示「尚無可用組織」。單一組織直接顯示名稱，多個組織可開啟選單。重新載入後以 GET 的結果恢復選取。
- 點選後暫停重複送出。成功時套用回應、關閉選單，手機同步收合側欄；清除前一組織的畫面資料並載入新組織內容。延遲抵達的舊請求不得覆蓋新畫面。
- 失敗時顯示錯誤並重新取得後端狀態；若成員資格已撤銷，依最新清單更新畫面。驗收涵蓋空清單、同名組織、鍵盤與手機切換、重新整理、失去資格與舊回應晚到。



[跨領域歸屬契約](../../architecture/contracts/organization-team-ownership.md) · [領域入口](README.md)
