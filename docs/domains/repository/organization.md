# Organization 所有權與 Team 歸屬

實作範圍：組織與 Team API、Firestore 寫入、側欄切換及團隊清單。文件路徑均位於 channels/{channelId} 下。現行群組資料與授權行為見[儲存庫規格](README.md)。

![Enterprise 與 Organization 分別管理 Team，Enterprise Team 可指派至多個旗下組織](diagrams/organization-team.svg)

[Mermaid 原始檔](diagrams/organization-team.mermaid)

## Enterprise（企業帳戶）

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `enterprises/{enterpriseId}` | 企業集中管理旗下組織與 Enterprise Team。Organization 可獨立存在，加入企業後以 enterpriseId 指向一個 Enterprise。 |
| 指派企業 Team | 具企業 Team 指派權的管理者指定旗下 Organization。後端核對 Team 的管理企業與目標組織的企業相同，再依組織與 Team 的組合鍵寫入；跨企業一律拒絕，重送不產生第二份名單。 |
| 移除指派 | 刪除指定組織的指派。依賴這筆指派的授權即失效，其他組織的指派與企業成員名單維持有效。 |

### 企業自助開通與共用頁面

| 操作 | 後端規則與前端呈現 |
|---|---|
| 顯示入口 | 組織 context 一併回傳 enterprises 與 canCreateEnterprise；空企業也對其 Owner 顯示。側欄直接使用回應。 |
| POST /api/v1/enterprises | 名稱為 1–120 字。後端於交易內核對登入者資格、建立企業、加入第一位 Owner 並記錄 enterpriseId。重送同一開通請求回傳既有企業。 |
| GET /api/v1/enterprises/{enterpriseId} | 企業 Owner 與管理者可看旗下組織及企業團隊清單；一般組織成員只看本人可見組織及已指派且通過 LINE 成員核驗的 Team。只有 Owner 收到 canManage=true。 |
| POST /api/v1/enterprises/{enterpriseId}/owners | 現任企業 Owner 可新增另一位 Owner；交易保留原有 Owner，支援既有 ownerUserId。 |
| POST /api/v1/enterprises/{enterpriseId}/organizations | 操作者須同時擁有企業與目標組織；已屬其他企業回 409。獨立組織可繼續單獨運作。 |

開通資格由服務端配置於同一渠道下的 `enterpriseEntitlements/{SHA-256(userId)}`（文件鍵使用專案 firestorekey.Digest）。設定 `enabled: true` 提供一次企業開通資格。建立成功後由後端保存 enterpriseId；缺少資格時直接呼叫建立 API 也回 403。頁面使用 OpenAPI 生成的型別，顯示後端回傳的清單與管理操作。

## Organization（組織）

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `organizations/{organizationId}` | 建立時由後端產生 UUID。每個組織恰有一位現任 Owner，唯一判定欄位為 `ownerUserId`。`name` 可修改，`createdBy` 保存建立者。 |
| 建立組織 | 已登入使用者建立 Organization，後端從已驗證身分填入 ownerUserId 與 createdBy。同一位使用者可建立多個組織。 |
| 移轉所有權 | 只有現任 Owner 能指定有效的現有成員為新 Owner。交易內重新核對現任 Owner 並更新 ownerUserId；舊 Owner 隨即失去 Owner 權限，createdBy 與 Team 歸屬維持原值。 |

## Team（團隊）

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `teams/{teamId}` | 文件 ID 等於 LINE groupId。首次綁定時成對寫入 managementScope（organization 或 enterprise）及 managementId（對應層級的 ID）。每個 Team 只有一個管理歸屬，舊群組可暫未綁定。 |
| 綁定群組 | 指定管理歸屬須具目標管理範圍的管理權及該群組的管理授權。交易內確認原歸屬；同一綁定可重送，不同管理者競爭時只有一方成功。可信 LINE 群組事件沿用既有流程自動建立 Team，不需人工核准；管理歸屬另行設定。 |
| `organizations/{organizationId}/enterpriseTeams/{teamId}` | 記錄一個 Enterprise Team 在此組織的指派。同一 Team 可指派給同企業的多個組織；成員名單由企業集中維護。 |

例如總部「資安組」可以用同一個 teamId 指派至台灣與日本兩個組織，各組織分別授予 Repository 權限。「台灣排班組」由台灣組織管理，只在台灣使用。後端判定資源存取時須核對目前的組織歸屬或有效指派、使用者資格及個別資源授權；既有 LINE 資料仍遵循原功能的成員核驗規則。

既有出勤與摘要繼續以原 groupId 查詢。既有群組在明確綁定前維持未歸屬；組織所有權本身不授予 LINE 群組資料讀取權。

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


### 管理操作與既有資料

| HTTP 操作 | 用途 |
|---|---|
| POST /api/v1/organizations | 登入者建立獨立組織。 |
| PUT /api/v1/organizations/{organizationId}/owner | 現任 Owner 移轉給組織現有成員。 |
| PUT /api/v1/teams/{teamId}/management | 為已自動建立的 Team 指定管理歸屬。 |
| PUT 或 DELETE /api/v1/organizations/{organizationId}/enterprise-teams/{teamId} | 新增或移除同企業內的 Team 指派。 |

組織的 memberUserIds 代表有效成員；管理歸屬操作讀取服務端維護的 managerUserIds。企業自助開通與組織加入流程見上方企業規格。既有 LINE 群組在下一筆可信事件時以 groupId 補齊 Team；事件只更新群組名稱與狀態，保留既有管理範圍與權限。


### LINE 一對一建立入口
![LINE 私訊建立流程](diagrams/private-chat-creation.svg)

[Mermaid 原始檔](diagrams/private-chat-creation.mermaid)

在私訊輸入「建立組織」或「建立企業」，五分鐘內回覆名稱即可建立；輸入「取消」結束。建立企業時，詢問名稱及實際建立時都檢查開通資格。已驗證的 webhook 決定建立者身分。成功回覆附上對應 MINI App 入口。對話狀態依渠道與使用者隔離；資源建立、結束對話及事件回覆紀錄在同一個 Firestore 交易完成，重送同一事件會取得原結果。群組事件沿用原有 Team 流程。
