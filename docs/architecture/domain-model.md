# Repository、Issue 與 Discussion 業務模型設計

狀態：後續開發設計參考，尚未核定完整產品規則；不構成現在實作、部署或資料遷移的要求。
整理日期：2026-11-12。檢查時 HEAD：`b24fce437621b757ef5d550e8fdd1b2bdbddbfd6`；工作目錄另有未提交變更，此 SHA 不代表全部工作內容。

## 文件歸屬與現況

設計文件與程式碼放在同一個 Git 儲存庫，跨業務模型集中於本頁，各業務維護自己的功能規格。
業務上的 Repository 指群組空間，與保存原始碼的 Git 儲存庫是不同概念。

- 命名政策只有一份：[命名規範](../development/naming.md#業務實體集合與-api-命名)。本頁保存三個主要業務的應用對照及後續資源模型參考。
- Repository 既有功能、授權及 REST 入口見 [群組 Repository](../features/repository.md)。本頁不取代其已實作規格。
- Issue 與 Discussion 是本次預留的業務設計；本頁未完成它們的實作盤點，不宣稱已有可用功能。
- GraphQL 欄位與 Connection 是未來介面設計，並非現有 API。現行 HTTP 來源仍為 [OpenAPI 契約](../../api/README.md)。

## 實體、查詢、集合與分頁型別

| 語意 | Repository | Issue | Discussion |
| --- | --- | --- | --- |
| 單一業務實體／未來 GraphQL Type | `Repository` | `Issue` | `Discussion` |
| 單一資源查詢欄位 | `repository` | `issue` | `discussion` |
| 多個資源的集合／集合查詢欄位 | `repositories` | `issues` | `discussions` |
| 集合的分頁型別 | `RepositoryConnection` | `IssueConnection` | `DiscussionConnection` |
| 分頁中的關聯項目型別 | `RepositoryEdge` | `IssueEdge` | `DiscussionEdge` |
| REST 集合名稱 | `/repositories` | `/issues` | `/discussions` |
| REST 單筆形式 | `/repositories/{id}` | `/issues/{id}` | `/discussions/{id}` |

上述 REST 名稱表達命名形式，Issue／Discussion 的版本前綴、是否巢狀於 Repository 及正式路由尚待契約設計。
既有 Repository API 為 `/api/v1/repositories` 與 `/api/v1/repositories/{id}`，不得依本頁直接改名。
`RepositoryConnection` 是含集合項目與分頁資訊的回傳包裝，不是單一 Repository，也不是新的資料庫 collection。
現有 `RepositoryPage` 等 REST DTO 維持原契約；不因預留 Connection 命名而重新命名。

## 業務邊界與關係

目標關係：一個 Repository 可容納多個 Issue 與 Discussion，每個 Issue／Discussion 歸屬一個 Repository。
建議以 `repositoryId` 表達歸屬；實際識別碼、唯一性、保存路徑與索引在實作前決定，不直接以群組 ID 取代 Repository ID。
這是邏輯關係，不代表必須使用 Firestore 子集合、外鍵或共用交易聚合。

| 業務 | 設計責任 | 必須保留的邊界 |
| --- | --- | --- |
| Repository | 容器識別、群組綁定、可見性與入口狀態 | 既有 Bot 管理模式維持；取得容器資格不自動授予所有子業務操作權 |
| Issue | 可追蹤事項及其生命週期 | 不直接套用出勤審核狀態；欄位、指派與完成規則待確認 |
| Discussion | 討論主題及其互動規則 | 分類、回覆、解答、鎖定與轉換 Issue 等能力待確認 |

Issue 與 Discussion 各自擁有狀態與操作；若未來需要互相連結或轉換，另訂識別、授權、原子性及追溯規則。
Repository 停用時子資源如何讀寫、保存多久、能否移轉或刪除，均須明確設計，不能推定連帶刪除。

## 未來 GraphQL 分頁設計

若未來採用 GraphQL，集合欄位回傳對應的 Connection；單筆欄位回傳對應實體。
例如 `repository(id: ...)` 回傳 `Repository`，`repositories(first: ..., after: ...)` 回傳 `RepositoryConnection`。
在單一 Repository 下，`issues` 與 `discussions` 分別回傳 `IssueConnection` 與 `DiscussionConnection`。
是否另提供根層 `issue`／`discussion` 查詢、未找到時的 null／錯誤語意，實作前於正式 schema 決定。

| 分頁元素 | 設計用途 |
| --- | --- |
| `edges` | 當頁的 Edge 集合，每筆含 `node` 與 opaque `cursor` |
| `node` | 單一 `Repository`、`Issue` 或 `Discussion` |
| `pageInfo` | 頁面資訊，初期可規劃 `endCursor` 與 `hasNextPage` |
| `first`／`after` | 正向分頁數量與游標；預設值、上限及錯誤語意待契約決定 |

游標應與查詢範圍、篩選及穩定排序一致，不能當成授權憑證。反向分頁、`nodes` 捷徑與 `totalCount` 按需求決定，不預先承諾。
既有 REST 可能回傳空頁但仍有 nextCursor；不得未經設計就把既有 nextCursor 等同 GraphQL endCursor 或套用不同的 hasNextPage 語意。
GraphQL 的 Connection 不要求 Domain Entity 或保存文件攜帶 edges／pageInfo；傳輸層負責映射。

## GitHub 資源模型參考目錄

以下僅保存後續建模素材與工程現場類比，不是已實作功能、排程或必建清單。
欄位位於特定父型別，不能視為全部存在於 GraphQL 根查詢；同名集合欄位可出現在不同父型別。
本表以固定 revision 的 GitHub schema 核對 Type 與欄位，並非要求自家服務沿用 GitHub 全部模型。

### 單一資源與集合欄位

| 單一資源 Type | 集合欄位 | GitHub 父型別範例 | 功能 |
| --- | --- | --- | --- |
| `Repository` | `repositories` | `RepositoryOwner`（如 `User`／`Organization`） | 儲存庫；本專案既有業務意義為群組空間 |
| `Issue` | `issues` | `Repository` | 問題、工作事項 |
| `Discussion` | `discussions` | `Repository` | 討論 |
| `Label` | `labels` | `Repository` | 標籤 |
| `Milestone` | `milestones` | `Repository` | 里程碑 |
| `Release` | `releases` | `Repository` | 發布版本 |
| `Deployment` | `deployments` | `Repository` | 部署紀錄 |
| `ProjectV2` | `projectsV2` | `Repository` | 專案管理 |
| `ProjectV2Item` | `items` | `ProjectV2` | 專案工作項目 |
| `DiscussionCategory` | `discussionCategories` | `Repository` | 討論分類 |
| `IssueType` | `issueTypes` | `Repository`／`Organization` | 議題類型 |
| `RepositoryTopic` | `repositoryTopics` | `Repository` | 儲存庫主題關聯 |
| `Team` | `teams` | `Organization` | 團隊 |

`ProjectV2` 的 V2 是 GitHub 的既有命名；本專案若需要專案管理，先決定自己的業務名稱與範圍，不據此預設存在 V1 或引入版本字尾。
`ProjectV2Item` 是專案中的項目，不必然等同 Issue；`RepositoryTopic` 表達儲存庫與主題的關聯，不直接當作標籤的別名。
這些集合在參考 schema 中使用對應 Connection，例如 `LabelConnection`、`ProjectV2ItemConnection`；是否在自家介面採用，依未來契約決定。

### 子資源與父資源欄位

| 子資源 Type | 父資源集合欄位 | 功能 |
| --- | --- | --- |
| `IssueComment` | `Issue.comments` | 問題留言 |
| `DiscussionComment` | `Discussion.comments` | 討論回覆 |
| `ReviewRequest` | `PullRequest.reviewRequests` | 審查請求 |
| `Reaction` | `Issue.reactions` | 表情反應 |
| `ReleaseAsset` | `Release.releaseAssets` | 發布附件 |

本表列出使用情境而非唯一歸屬；例如 Reaction 可用於其他支援反應的內容。
`PullRequest` 在此作為 GitHub 參考父型別，不代表本專案計畫實作程式碼合併功能。
父子關係不直接決定 REST 巢狀路由或資料庫保存位置；仍須明訂存取範圍與刪除規則。

### 資源模型與工程現場應用類比

| 資源模型 | GitHub 功能 | 工程現場應用範例 |
| --- | --- | --- |
| `User` | 使用者身分，可參與操作；部分資源可由使用者擁有 | 使用 LINE 登入識別個人 |
| `Organization` | 管理儲存庫、團隊與組織資源 | 工程公司管理多個專案 |
| `Team` | 組織內的成員分組與協作單位 | 工程部、採購部、施工班組 |
| `IssueComment` | Issue 下方的留言、回報與討論紀錄 | 施工人員回報「材料已到場」 |
| `DiscussionComment` | Discussion 內的回覆，支援對話與問答互動 | 回覆施工方案討論、補充意見 |
| `Reaction` | 對內容做快速反應 | 表示知悉、支持或贊同 |
| `PullRequestReview` | 正式審查 Pull Request，可批准或要求修改 | 借鑑為文件或變更單的審核決策 |
| `ReviewRequest` | 請求指定對象進行審查 | 指派工程經理審核變更申請 |

身分識別、組織歸屬、容器成員與業務授權是不同責任；LINE 登入不自動代表公司成員或資源擁有者。
Organization／Team 是否需要獨立實體、如何驗證成員及如何與現有群組空間對應，須另行產品設計，不改變目前 Repository 的 Bot 管理模式。
`ReviewRequest` 是待審任務或指派，`PullRequestReview` 是審查紀錄／決策；兩者不互相替代。
Reaction、留言或被指派審查均不等於正式核准，也不自動取得核准權。
工程變更單若採用此概念，應取符合領域的名稱並訂定自己的申請、審核與生效狀態，不直接複製 GitHub 型別或出勤審核規則。

### Projects V2 型別與欄位模型

本節延伸專案管理參考，全部屬於後續開發素材；GitHub 型別名不直接成為本專案的實體、API 或資料集合。

| GitHub 型別 | 種類與責任 | 工程現場應用參考 |
| --- | --- | --- |
| `ProjectV2` | 專案本體，包含工作項目、欄位與視圖 | 整合某工程的工作規劃 |
| `ProjectV2Item` | 專案內的一筆工作項目；與引用的內容分開識別 | 將施工事項納入專案追蹤 |
| `ProjectV2Field` | 一般欄位定義物件，不是所有欄位的共同父物件 | 定義施工區域、預計日期等欄位 |
| `ProjectV2SingleSelectField` | 單選欄位定義，包含選項設定 | Priority／優先級、工班或狀態 |
| `ProjectV2IterationField` | 週期欄位定義，管理 iteration 設定 | 每週施工工作規劃 |
| `ProjectV2ItemFieldValue` | 欄位值的 **union**，依實際型別讀取日期、文字、單選或週期等值 | 某筆工作的預計日期、優先級或施工週次 |
| `ProjectV2View` | 已保存的專案視圖設定 | 每日工作、各工班工作或工程時間軸 |

共同欄位定義介面為 `ProjectV2FieldCommon`；`ProjectV2FieldConfiguration` 是包含一般、單選、週期等欄位定義的 union。
共同欄位值介面為 `ProjectV2ItemFieldValueCommon`，不可將它與 `ProjectV2ItemFieldValue` union 混稱。
欄位定義、單選選項與每筆項目的欄位值是不同資料；修改某筆工作的 Priority 不等於修改整個專案的選項設定。
此處只列需求涉及的型別，並非 GitHub 全部欄位種類清單。

`ProjectV2Item.content` 使用 `ProjectV2ItemContent` union，可引用 `Issue`、`PullRequest` 或 `DraftIssue`。
項目 ID 與內容 ID 不互換；同一事項納入不同專案時，專案欄位值的歸屬須分別處理。
GitHub 的 Issue／PR 指派、標籤等來源屬性與專案自訂欄位責任不同，不應一律當成項目自訂值更新。
本專案是否支援草稿工作、跨 Repository 收集及同一事項加入多個專案，仍待產品規格確認；PR 僅作參考，不新增程式碼合併需求。

### Projects 功能與工程現場用途

以下為 GitHub Projects 的產品能力分類，並非每一列都有同名 GraphQL Type 或完整可寫 API；採用前需個別核對。

| 能力 | 功能 | 工程現場應用範例 |
| --- | --- | --- |
| Items | 收集 Issue、PR 或 Draft Issue | 整合各項施工工作 |
| Table | 表格模式管理項目與欄位 | 工作清單、負責人、進度 |
| Board | Kanban 看板，依欄位值分欄 | 待辦、進行中、完成 |
| Roadmap | 依日期或週期呈現時間軸 | 工程排程、施工階段 |
| Custom Fields | 設定自訂欄位及其型別 | 施工區域、優先級、預計日期 |
| Views | 保存多組篩選與呈現設定 | 每日工作、各工班工作 |
| Filtering | 依欄位與內容屬性篩選 | 只看逾期、未完成事項 |
| Grouping | 依支援欄位分組 | 依工班、狀態或負責人分類 |
| Iteration | 以可配置長度與休息區間的週期規劃工作 | 每週施工工作規劃 |
| Workflows | 自動化項目新增、狀態更新或歸檔等管理 | Issue 完成時更新專案狀態 |
| Insights | 圖表與進度分析 | 查看各階段完成情況 |
| Permissions | 專案存取與操作控制 | 管理誰可查看、編輯或管理專案 |
| Templates | 重複使用專案設定 | 建立標準工程工作看板 |

Table、Board 與 Roadmap 是同一批專案項目的不同呈現方式，不各自複製一套工作資料。
Views 保存布局、篩選、分組與排序等設定；Filtering／Grouping 是查詢與呈現能力，不因一個項目被多個視圖選中而複製它。
Roadmap 顯示日期或週期不代表具備工序相依、關鍵路徑或自動排程能力；這些需求須另訂規格。
Iteration 是規劃區間，不等同實際出勤、工時結算或施工完成證明。

### Projects 採用邊界與待決策

- **歸屬與引用**：先決定專案由 Organization、Repository 或其他工程容器擁有，以及跨容器項目的引用與刪除規則；不能由 Repository 上可查詢 `projectsV2` 就推定專案只屬於單一 Repository。
- **狀態**：Issue 的生命週期與專案 Status 欄位分別建模；完成、取消、封存及移出專案的效果另訂，不能視為同一操作。
- **權限**：分開定義專案可見性、項目來源可見性、欄位編輯與管理權；專案存取不自動授予來源事項的讀寫或審核權，篩選與隱藏欄位也不能充當授權。
- **自動化**：Workflows 若會更新狀態，需明訂觸發條件、執行身分、授權、重送、失敗及 Audit；自動移至「完成」不代表已通過施工驗收或業務核准。
- **分析**：Insights 的完成率、分母、時區與歷史資料來源須先定義，不能只靠目前狀態推算過去進度。
- **範本**：決定 Templates 複製哪些欄位、選項、視圖、項目與自動化，以及建立時的權限；不預設連帶複製成員、正式工作或敏感內容。
- **後續文件**：確定採用後建立專案管理功能規格，按操作模型補齊六個問題、Context／Authority／Decision 與必要 Audit，再設計契約及驗收；本頁不新增資料保存路徑或正式 API。

### 採用前待確認

- 先確認使用者需求，再選擇需要的模型；每個被採用的業務建立功能文件與操作契約。
- 明訂擁有者、Repository／Organization 範圍、成員資格與每個操作的授權來源。
- 確認識別碼、必填欄位、狀態、父子關係、跨容器引用及保存／刪除政策。
- 審核另訂指派、決策權、本人迴避、撤回、重審、併發與稽核；附件另訂上傳、存取及保存規則。
- 完成契約與驗收後才更新實作狀態；本參考目錄本身不新增 API、資料集合或操作權限。

## 後續開發文件與實作順序

1. Repository 維持 [既有功能文件](../features/repository.md)；Issue 與 Discussion 開發時分別建立 `docs/features/issues.md`、`docs/features/discussions.md`，並加入功能索引。
2. 功能文件先按 [操作契約範本](business-operation.md#功能操作契約範本) 補齊每個操作：誰操作、對象、行為、權限、業務條件及成功結果；Context、Authority、Decision 與必要 Audit 明確記錄。
3. 開發前確認 Issue／Discussion 的建立、讀取、修改與生命週期操作，以及權限來源、可見範圍、併發版本、重送、失敗及稽核語意。不從 GitHub 功能自行推定產品需求。
4. HTTP 契約依業務放在 `api/contract/<業務>/`，由既有 OpenAPI 入口組合；僅在採用 GraphQL 的決策成立後規劃 schema 入口，不先建立空契約或生成碼。
5. 程式按業務維護，模組 README 連回功能規格；契約來源、設計說明與實作避免重複維護相同欄位定義。
6. 完成契約相容性、權限隔離、分頁及所選操作的驗收；文件狀態依實際證據更新，不能因本設計存在就標為已實作。

## 參考與適用範圍

- [GitHub GraphQL fpt 資料](https://github.com/github/docs/tree/528435f1d245ebc2a1c2a8d1c0254a5d4ad17b21/src/graphql/data/fpt)：參考 `schema.docs.graphql`、業務分類的 `schema-issues.json`／`schema-discussions.json` 及 `category-map.json`；借鏡類型描述與分類，不複製 GitHub 產品權限與狀態。2026-11-12 另直接核對此固定 revision 的 `schema.docs.graphql`，確認資源參考表的型別、父欄位與 Connection 回傳型別；不宣稱它是最新 GitHub API。
- [GraphQL 官方分頁說明](https://graphql.org/learn/pagination/)：2026-11-12 透過 Context7 `/websites/graphql_learn` 查證 Connection、Edge、cursor、PageInfo 與 first／after 範例；該 Learn 來源未提供固定版本。本頁選用單數實體加 Connection 的名稱是專案慣例，不宣稱所有 GraphQL 集合都強制如此命名。
- [GitHub Projects 文件](https://docs.github.com/en/issues/planning-and-tracking-with-projects)：2026-11-12 經 Context7 `/github/docs` 查閱 iteration、視圖、篩選及 API 更新項目欄位說明；Context7 來源為 main，未固定 revision。Projects 型別種類另以本頁上述固定 revision 的 `schema.docs.graphql` 核對，包含 `ProjectV2ItemContent`、欄位定義及欄位值 union／共同介面；未查證所有 Workflows、Insights、Permissions、Templates 的方案限制或 API 覆蓋，採用時仍須核對。
- 目前只完成文件設計與本機文件檢查；未建立 GraphQL runtime、Issue／Discussion API、資料 collection 或遷移。
