# 個人或工程群組擁有的 Project

## 邊界與身份

個人或工程群組擁有 Project，並綁定 Repository。Project 擁有專案內容、收錄項目、版本、作者管理權及防重規則；Repository 是物件／子物件／問題／討論的來源讀取範圍，不再是 Project 的聚合根。建立時選擇同一 owner 的初始來源，之後可連結其他有權讀取的 Repository。

工程群組直接保存於 `channels/{channelId}/groups/{groupId}`，由已驗簽的群組生命週期獨立建立／更新，候選索引及目前成員核驗獨立於 Repository Service。

Repository 的既有 ID、groupId 綁定、事項與討論保存位置不變。初始 Repository 的 ownerId 必須等於 Project 的 OwnerID；後續來源依各自群組授權；從 sourceRepositoryId 不得推導新增管理權或跳過當下來源授權。

## 操作契約

| 欄位 | Group 列表／詳情 | Project 列表／詳情／建立／更新 | ProjectItem 列表／加入／移除 |
| --- | --- | --- | --- |
| Actor | 既有 Session／Bearer 驗證的本人 | 同左 | 同左 |
| Resource／Action | channel 內 Group；List／Get | owner 內 Project；List／Get／Create／Update | Project 自有項目；List／Add／Remove |
| Context／Authority | 可信群組候選不代表授權；ACTIVE Group 與當下群組成員核驗 | 個人本人或 Group 成員可讀／建立，Project 建立者管理內容；來源資格另外核驗 | Project 建立者管理收錄，來源事項仍依目前 Repository 讀取資格 |
| Policy／Decision | 非成員、停用、不存在不揭露；外部核驗失敗不當作允許 | 一個初始來源及多個可連結來源；沿用內容限制、expectedVersion 與防重鍵；Project 無獨立完成狀態 | 同 Project 不重複，同事項可在不同 Project；項目 ID 與 Issue ID 分開，來源資格不允許則不揭露來源內容 |
| Result | 本人可開啟的 Group／基本資料，不授予 Issue 寫權 | Project 自有資料／版本；不寫 Repository、Issue 內容或狀態 | 僅改關聯及 Project 版本，來源事項完全不變；進度由可讀來源查得，不複製 Issue 狀態 |
| 重送／併發 | 讀取無副作用；外部成員資格不宣稱與資料庫原子 | 舊建立鍵及收據維持同結果／同鍵異內容衝突；更新過期版本拒絕，無變化不加版本 | expectedVersion 序列化收錄／移除，結果不明先讀取再決定 |
| Event／Audit | 保存必要群組事件排序，不記 token／profile | 保留原作者、建立／更新時間及版本追溯 | 關聯變動與 Project metadata 同交易，無外部通知 |

Project 自有資料的授權不得再要求 Repository 作為父容器。來源成員資格被撤銷或來源不可用時，來源 Issue 的內容、狀態及衍生進度不能由 Project 權限繞過；不得把未知統計回傳成零或以舊來源資料冒充成功。前端隱藏按鈕不能代替後端檢查。

Canonical來源hydration失敗時，已授權的Project自有metadata仍可讀取／管理，回應以`sourceAvailable=false`標明進度不可用；來源項目內容不因此開放。Project自有container、migration marker、item結構錯誤以及取消／逾時仍回錯誤，不當成來源降級；legacy路徑維持原失敗語義。[第二輪收斂證據](../../reports/2026-11-13/architecture-convergence-pass2-20261113.md)記錄此故障邊界與實際驗證。

## API、頁面與側欄

- Group 提供獨立 List／Get；Project 新入口為 `/api/v1/owners/{ownerId}/projects` 及其專案／items 子資源。Create 明確指定 `defaultRepositoryId`，URL 的 Group 或 Project ID 不是授權憑證。
- 舊 `/api/v1/repositories/{repositoryId}/projects...` 保留相容 adapter，解析來源及 Group 後呼叫同一組用例，不維護第二份授權或寫入邏輯。原回應、ID、版本、防重及分頁契約依相容檢查保留。
- 舊Repository統計的projects欄位相容回報以該Repository為來源的專案數，依唯一保存位置讀取；這是唯讀來源關聯統計，不表示Repository擁有或可寫Project。
- 正式專案頁為 `/miniapp/projects`，URL 保存 Group／Project 選取；側欄「專案」與「儲存庫」同層。Project 頁透過 Group API 與本人 Session 管理資料，不需要 RepositoryProvider 作為 Project 生命週期。
- Repository不再提供專案頁籤，專案由獨立入口開啟。新增專案的來源選擇只列可讀來源，不能因此取得事項修改／審核權。切換 Group、Project、身份或登出時取消舊請求並隔離畫面資料，上一頁、重新整理與直接連結可回到同一目標。
- 原官方 shadcn primitives 及 sidebar-07 模板維持不變；業務導覽在應用層組合。Pages Router 維持靜態輸出，Go 靜態頁路由亦註冊新入口，不新增前端框架或依賴。

## 最小來源引用骨架

專案只收錄來源引用，不建立第二份物件、問題或討論資料；每個 Project 可連結多個來源庫。新增`references`入口支援object、subobject、issue、discussion四類，沿用同一Project與items集合。舊`items`入口只回Issue欄位／狀態，既有Issue引用ID、移除與IssueCount／CompletedIssueCount語意保留，不將物件完成數混成專案百分比。

新引用只保存sourceType、sourceId及子物件必要的parentObjectId與關聯追溯欄位；跨庫引用保存 sourceRepositoryId。物件改名、數量或狀態更新後，列表重新讀取來源目前值，不儲存一份可分歧的名稱／狀態快照。子引用需同時驗父物件路徑，不能只有子ID。來源引用不是存取或修改來源的授權。

| 欄位 | 列出引用 | 加入引用 | 移除引用 |
| --- | --- | --- | --- |
| Actor／Resource／Action | 本人可信Session／Bearer；owner 內 Project的來源引用；List | 同Project內Add | 同Project內Remove |
| Context／Authority | 個人本人或 ACTIVE Group 的目前成員；來源讀權另外核驗 | 同左且Project建立者；已連結來源庫、ACTIVE 與該庫的當下讀權 | Project建立者及Project版本，不要求來源仍可讀 |
| Policy／Decision | 來源缺失／失權保留占位，sourceUnavailable且不提供舊標題、數量或狀態；不令整個列表迴圈失敗 | sourceType與ID合法、父子歸屬正確，同Project同來源不重複 | 只刪引用，不刪來源，不繞過Project管理權或版本 |
| Result | 最新來源唯讀preview；不改來源與Project | 交易保存引用並更新Project版本／audit | 交易移除引用並更新Project版本／audit，來源不變 |
| 重送／併發 | 分頁游標綁本人與Project scope | 決定性來源鍵綁類型、父與來源ID，重送已存在同引用不重複；其他過期版本變更409 | 同引用重複移除不刪其他資料；其他過期版本變更409 |

四種引用共用 items 文件格式，將來源庫、來源類型、父物件及來源 ID 納入決定性鍵；不同庫的同名識別可獨立收錄。

新增API為GET／POST `/api/v1/owners/{ownerId}/projects/{id}/references`及DELETE同路徑`/{itemId}`；權威schema及生成型別隨本輪同步。加入時提交 sourceRepositoryId、sourceType、sourceId、可選 parentObjectId 與 expectedVersion。Project PATCH 可帶 linkRepositoryId 或 unlinkRepositoryId；解除後保留引用占位。

畫面保留現有專案建立／編輯，可選擇與連結儲存庫，再選種類再從來源清單選擇，子物件先選父再選子，不要求輸入UUID。來源失權仍顯示不可存取占位；Project建立者可移除引用，不能用快取舊preview洩漏內容。加入／移除待處理或結果未知時阻止切換Group／Project及卸載，先核回目前版本與引用再決定續送。

## 保存路徑與顯示範圍

核心模型以 `OwnerID` 表示個人或工程群組。HTTP 使用 `owners/{ownerId}`，Firestore 依身分使用 `users/{userId}` 或 `groups/{groupId}`，Project 文件欄位為 `ownerId`，核心模型以 `DefaultRepositoryID` 表示建立時的預設庫（API 與 Firestore 欄位 `defaultRepositoryId`）。每筆 ProjectReference 各自保存所引用的來源庫。

Project 固定保存於 `channels/{channelId}/owners/{ownerId}/projects/{projectId}`。Repository 路由先解析工程群組，再存取同一個 Project。初始來源只作預設選擇，其他有效來源保存在 linkedRepositories。建立收據沿用工程群組 內的 projectCreationScopes；舊 projectStorage marker 不參與服務端讀寫。舊遷移工具僅供離線歷史資料處理，操作前須停止所有讀寫服務。

工程對象、待處理問題、決策討論分區顯示；專案詳情不把預設庫的 Issue 計數顯示為跨庫進度。Repository 統計中的專案數指以該庫為初始來源的專案數。

## 歷史版本驗證（2026-11-13，含舊保存路由）

- 已維護Space service／Firestore及Project service／Firestore測試，另有migration CLI單元與emulator測試。隔離本機emulator的race驗證涵蓋無Repository的Space事件／候選、owner與來源撤權分離、非成員／作者／版本／同Space限制、canonical與legacy的同ID／收據重送、原Issue不變、marker鎖與部分copy不可見，以及rollback後新頁仍可讀legacy資料。
- 遷移實測dry-run零寫入、writer-ready門檻、完整items／changes／收據及時間版本保存、異內容碰撞／來源及目標漂移拒絕、無漂移回復與新寫入後不切回舊資料；舊Repository統計在移轉前後不重複計數。
- API能力檢查、生成同步及相容性驗證通過（61項操作、27個契約來源）；Go格式、lint、module verify、vet、完整既有race suite、build與漏洞檢查通過。漏洞檢查無應用可達項，仍有3筆未被應用呼叫的required-module項目，不稱所有依賴無漏洞。
- 前端型別及29個現有框架測試通過，包含新增6個Space／Project／身份與來源隔離、同層側欄、URL回復、原鍵重試及來源不可用情境；Next靜態build輸出新projects頁。官方UI primitives、use-mobile及sidebar-07模板與基準完全相同，未新增依賴。

LINE官方來源庫本次HEAD為 `e39b8daf919dc7a9396e44b5fc55a2aad918d932`，核對group-chats與其成員profile引用。Context7核對Next `16.4.5` Pages Router，並對照安裝版；Firestore官方Go client交易來源與本機 `1.29.2` 核對讀取先於寫入及callback可重試。Google Developer Knowledge／Google skills未提供可用入口，未宣稱其主機驗證或正式服務驗收。實際程式來源先用Graphify與Repomix定位／完整交叉核對，修改後重建並驗證圖與來源涵蓋；AST關係與安全掃描不取代執行測試。

本次僅對隔離emulator執行apply／rollback，沒有對正式資料庫執行遷移、清理或部署。正式環境須先確認writer升級與namespace、審閱dry-run manifest，依操作指南另行執行；有canonical新寫入的回復仍需額外copy-back方案，不在本工具能力內。

## 工程驗收送件

有 Project 所屬群組 成員資格、且可讀取已連結來源庫的使用者，可對 Object 或 SubObject 引用送驗。GET／POST `/api/v1/owners/{ownerId}/projects/{id}/references/{itemId}/inspection-submissions` 讀取紀錄或提交檢查說明。送件保存於引用下的 inspectionSubmissions，記錄來源版本、送件人與時間；施工進度維持原值。同一引用一次只能有一筆待審，同人重送相同說明回傳原紀錄。有送件紀錄的引用不可移除。

查看歷史需通過協作空間與來源庫授權、庫連結及狀態檢查；原物件被刪除仍可讀取已保存的送件。新送件必須能讀取目前有效的物件與版本。
