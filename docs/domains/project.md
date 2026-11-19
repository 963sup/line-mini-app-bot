# 個人或工程群組擁有的-project

<a id="個人或工程群組擁有的-project"></a>

<a id="project-item-progress"></a>

## 專案項目進度

每個 Project item 各自保存待辦、進行中或完成。缺少欄位的既有項目讀為待辦。來源狀態依[問題生命週期](repository/issues.md#issue-lifecycle)、工程物件或巡檢規則讀取。畫面分別標示「本專案進度」與「來源狀態」。更新進度沿用 Project 編輯權、版本檢查與交易；來源讀權仍由來源庫核驗。

驗收時，將同一 Issue 加入維修與稽核兩個 Project。完成維修項目後重新載入兩個看板，只有維修項目顯示完成，稽核項目保留原進度，來源 Issue 仍為開啟。關閉來源 Issue 後，兩個看板的來源狀態都應反映關閉，各自的專案進度保留。

<a id="邊界與身份"></a>

```mermaid
flowchart LR
  h["邊界與身份"]
  h --- n0["個人或工程群組擁有 Project，並綁定 <br/>Repository。Project 擁有專案內容、收錄項目、<br/>版本、作者管理權及防重規則；Repository 是物<br/>件／子物件／問題／討論的來源讀取範圍，不再是<br/> Project 的聚合根。建立時選擇同一 owner 的初<br/>始來源，之後可連結其他有權讀取的 Repository<br/>。"]
  h --- n1["工程群組直接保存於 channels/{channelId}<br/>/groups/{groupId}，由已驗簽的群組生命週期獨<br/>立建立／更新，候選索引及目前成員核驗獨立於 <br/>Repository Service。"]
```

<a id="操作契約"></a>

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["欄位：Actor<br/>Group 列表／詳情：既有 Session／Bearer 驗證<br/>的本人<br/>Project 列表／詳情／建立／更新：同左<br/>ProjectItem 列表／加入／移除：同左"]
  h --- n1["欄位：Resource／Action<br/>Group 列表／詳情：channel 內 Group；List／<br/>Get<br/>Project 列表／詳情／建立／更新：owner 內 <br/>Project；List／Get／Create／Update<br/>ProjectItem 列表／加入／移除：Project 自有項<br/>目；List／Add／Remove"]
  h --- n2["欄位：Context／Authority<br/>Group 列表／詳情：可信群組候選不代表授權；<br/>ACTIVE Group 與當下群組成員核驗<br/>Project 列表／詳情／建立／更新：個人本人或 <br/>Group 成員可讀／建立，Project 建立者管理內容<br/>；來源資格另外核驗<br/>ProjectItem 列表／加入／移除：Project 建立者<br/>管理收錄，來源事項仍依目前 Repository 讀取資<br/>格"]
  h --- n3["欄位：Policy／Decision<br/>Group 列表／詳情：非成員、停用、不存在不揭露<br/>；外部核驗失敗不當作允許<br/>Project 列表／詳情／建立／更新：一個初始來源<br/>及多個可連結來源；沿用內容限制、<br/>expectedVersion 與防重鍵；Project 無獨立完成<br/>狀態<br/>ProjectItem 列表／加入／移除：同 Project 不<br/>重複，同事項可在不同 Project；項目 ID 與 <br/>Issue ID 分開，來源資格不允許則不揭露來源內<br/>容"]
  h --- n4["欄位：Result<br/>Group 列表／詳情：本人可開啟的 Group／基本資<br/>料，不授予 Issue 寫權<br/>Project 列表／詳情／建立／更新：Project 自有<br/>資料／版本；不寫 Repository、Issue 內容或狀<br/>態<br/>ProjectItem 列表／加入／移除：僅改關聯及 <br/>Project 版本，來源事項完全不變；進度由可讀來<br/>源查得，不複製 Issue 狀態"]
  h --- n5["欄位：重送／併發<br/>Group 列表／詳情：讀取無副作用；外部成員資格<br/>不宣稱與資料庫原子<br/>Project 列表／詳情／建立／更新：舊建立鍵及收<br/>據維持同結果／同鍵異內容衝突；更新過期版本拒<br/>絕，無變化不加版本<br/>ProjectItem 列表／加入／移除：<br/>expectedVersion 序列化收錄／移除，結果不明先<br/>讀取再決定"]
  h --- n6["欄位：Event／Audit<br/>Group 列表／詳情：保存必要群組事件排序，不記<br/> token／profile<br/>Project 列表／詳情／建立／更新：保留原作者、<br/>建立／更新時間及版本追溯<br/>ProjectItem 列表／加入／移除：關聯變動與 <br/>Project metadata 同交易，無外部通知"]
```

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["Project 自有資料的授權不得再要求 Repository <br/>作為父容器。來源成員資格被撤銷或來源不可用時<br/>，來源 Issue 的內容、狀態及衍生進度不能由 <br/>Project 權限繞過；不得把未知統計回傳成零或以<br/>舊來源資料冒充成功。前端隱藏按鈕不能代替後端<br/>檢查。"]
  h --- n1["Canonical來源hydration失敗時，已授權的<br/>Project自有metadata仍可讀取／管理，回應以<br/>sourceAvailable=false標明進度不可用；來源項<br/>目內容不因此開放。Project 容器或 item 結構錯<br/>誤以及取消／逾時仍回錯誤，不當成來源降級；<br/>legacy路徑維持原失敗語義。"]
```

<a id="api頁面與側欄"></a>

```mermaid
flowchart LR
  h["API、頁面與側欄"]
  h --- n0["Group 提供獨立 List／Get；Project 新入口為 <br/>/api/v1/owners/{ownerId}/projects 及其專案／<br/>items 子資源。Create 明確指定 <br/>defaultRepositoryId，URL 的 Group 或 Project<br/> ID 不是授權憑證。"]
  h --- n1["舊 /api/v1/repositories/{repositoryId}<br/>/projects... 保留相容 adapter，解析來源庫及<br/>其擁有者 後呼叫同一組用例，不維護第二份授權<br/>或寫入邏輯。原回應、ID、版本、防重及分頁契約<br/>依相容檢查保留。"]
  h --- n2["舊Repository統計的projects欄位相容回報以該<br/>Repository為來源的專案數，依唯一保存位置讀取<br/>；這是唯讀來源關聯統計，不表示Repository擁有<br/>或可寫Project。"]
  h --- n3["正式專案頁為 /miniapp/projects，URL 保存 <br/>owner／Project 選取；側欄「專案」與「儲存庫<br/>」同層。Project 頁透過 Group API 與本人 <br/>Session 管理資料，不需要 RepositoryProvider <br/>作為 Project 生命週期。"]
  h --- n4["Repository不再提供專案頁籤，專案由獨立入口開<br/>啟。新增專案的來源選擇只列可讀來源，不能因此<br/>取得事項修改／審核權。切換 Group、Project、<br/>身份或登出時取消舊請求並隔離畫面資料，上一頁<br/>、重新整理與直接連結可回到同一目標。"]
  h --- n5["原官方 shadcn primitives 及 sidebar-07 模板<br/>維持不變；業務導覽在應用層組合。Pages Router<br/> 維持靜態輸出，Go 靜態頁路由亦註冊新入口，不<br/>新增前端框架或依賴。"]
```

<a id="最小來源引用骨架"></a>

## 最小來源引用骨架

現行實作

```mermaid
erDiagram
  PROJECT ||--o{ SOURCE_REFERENCE : 收錄
  REPOSITORY ||--o{ SOURCE_REFERENCE : 提供來源
```

```mermaid
flowchart LR
  h["最小來源引用骨架"]
  h --- n0["每個 Project 可連結多個來源庫。引用指向物件<br/>、子物件、事項或討論，四類共用 items 集合。<br/>舊 items 入口保留 Issue 欄位、狀態、引用 ID<br/>、移除及 IssueCount／CompletedIssueCount 語<br/>意。物件完成數與專案事項統計分開。"]
  h --- n1["新引用只保存sourceType、sourceId及子物件必要<br/>的parentObjectId與關聯追溯欄位；跨庫引用保存<br/> sourceRepositoryId。物件改名、數量或狀態更<br/>新後，列表重新讀取來源目前值，不儲存一份可分<br/>歧的名稱／狀態快照。子引用需同時驗父物件路徑<br/>，不能只有子ID。來源引用不是存取或修改來源的<br/>授權。"]
```

```mermaid
flowchart LR
  h["最小來源引用骨架"]
  h --- n0["欄位：Actor／Resource／Action<br/>列出引用：本人可信Session／Bearer；owner 內 <br/>Project的來源引用；List<br/>加入引用：同Project內Add<br/>移除引用：同Project內Remove"]
  h --- n1["欄位：Context／Authority<br/>列出引用：個人本人或 ACTIVE Group 的目前成員<br/>；來源讀權另外核驗<br/>加入引用：同左且Project建立者；已連結來源庫<br/>、ACTIVE 與該庫的當下讀權<br/>移除引用：Project建立者及Project版本，不要求<br/>來源仍可讀"]
  h --- n2["欄位：Policy／Decision<br/>列出引用：來源缺失／失權保留占位，<br/>sourceUnavailable且不提供舊標題、數量或狀態<br/>；不令整個列表迴圈失敗<br/>加入引用：sourceType與ID合法、父子歸屬正確，<br/>同Project同來源不重複<br/>移除引用：只刪引用，不刪來源，不繞過Project<br/>管理權或版本"]
  h --- n3["欄位：Result<br/>列出引用：最新來源唯讀preview；不改來源與<br/>Project<br/>加入引用：交易保存引用並更新Project版本／<br/>audit<br/>移除引用：交易移除引用並更新Project版本／<br/>audit，來源不變"]
  h --- n4["欄位：重送／併發<br/>列出引用：分頁游標綁本人與Project scope<br/>加入引用：決定性來源鍵綁類型、父與來源ID，重<br/>送已存在同引用不重複；其他過期版本變更409<br/>移除引用：同引用重複移除不刪其他資料；其他過<br/>期版本變更409"]
```

```mermaid
flowchart LR
  h["最小來源引用骨架"]
  h --- n0["四種引用共用 items 文件格式，將來源庫、來源<br/>類型、父物件及來源 ID 納入決定性鍵；不同庫的<br/>同名識別可獨立收錄。"]
  h --- n1["新增API為GET／POST /api/v1/owners/{ownerId}<br/>/projects/{id}/references及DELETE同路徑/{<br/>itemId}；權威schema及生成型別隨本輪同步。加<br/>入時提交 sourceRepositoryId、sourceType、<br/>sourceId、可選 parentObjectId 與 <br/>expectedVersion。Project PATCH 可帶 <br/>linkRepositoryId 或 unlinkRepositoryId；解除<br/>後保留引用占位。"]
  h --- n2["畫面保留現有專案建立／編輯，可選擇與連結儲存<br/>庫，再選種類再從來源清單選擇，子物件先選父再<br/>選子，不要求輸入UUID。來源失權仍顯示不可存取<br/>占位；Project建立者可移除引用，不能用快取舊<br/>preview洩漏內容。加入／移除待處理或結果未知<br/>時阻止切換Group／Project及卸載，先核回目前版<br/>本與引用再決定續送。"]
```

<a id="保存路徑與顯示範圍"></a>

```mermaid
flowchart LR
  h["保存路徑與顯示範圍"]
  h --- n0["位置：HTTP<br/>Project 路徑：/api/v1/owners/{ownerId}<br/>/projects/{id}"]
  h --- n1["位置：Firestore 個人專案<br/>Project 路徑：channels/{channelId}/users/{<br/>userId}/projects/{projectId}"]
  h --- n2["位置：Firestore 群組專案<br/>Project 路徑：channels/{channelId}/groups/{<br/>groupId}/projects/{projectId}"]
```

```mermaid
flowchart LR
  h["保存路徑與顯示範圍"]
  h --- n0["Project 文件保存 ownerId 與 <br/>defaultRepositoryId。同一擁有者底下的 <br/>projectCreationScopes 保存建立收據，Project <br/>底下的 linkedRepositories 保存庫連結；每筆引<br/>用以 sourceRepositoryId 定位來源。"]
  h --- n1["工程對象、待處理問題、決策討論分區顯示；專案<br/>詳情不把預設庫的 Issue 計數顯示為跨庫進度。<br/>Repository 統計中的專案數指以該庫為初始來源<br/>的專案數。"]
```

<a id="工程驗收送件"></a>

## 工程驗收送件

現行實作

```mermaid
flowchart TB
  A["核驗個人擁有者或群組當下成員、庫連結、狀態與來源讀權"] --> B{"讀取歷史或新送件？"}
  B -->|歷史| H["讀取已保存送件，原物件已刪除仍可查"]
  B -->|新送件| O["核驗 Object 或 SubObject 引用及目前有效來源版本"]
  O --> R{"同人重送相同說明？"}
  R -->|是| X["回原紀錄"]
  R -->|否| P{"此引用已有待審送件？"}
  P -->|有| F["維持既有待審送件"]
  P -->|無| S["保存說明、來源版本、送件人與時間"]
```

```mermaid
flowchart LR
  h["工程驗收送件"]
  h --- n0["GET／POST /api/v1/owners/{ownerId}/projects/<br/>{id}/references/{itemId}<br/>/inspection-submissions 分別讀歷史與提交。資<br/>料保存在引用下的 inspectionSubmissions。施工<br/>進度維持原值，有送件紀錄的引用不可移除。"]
```
