# Space 擁有的獨立 Project

## 邊界與身份

Project 與 Repository 是 Space 內的平行業務。Project 擁有專案內容、收錄項目、版本、作者管理權及防重規則；Repository 僅是來源事項的讀取範圍，不再是 Project 的聚合根。首版每個 Project 只選一個來源 Repository，不建立跨來源框架、自訂狀態、欄位或視圖。

既有程式沒有獨立 Space 實體；本次明訂 Space 對應同一 Messaging API channel 內的 LINE 群組。`spaceId = groupId`，channel 作外層隔離；這是沿用既有可信群組識別的明確身份策略，不將隨機 Repository ID 改名為 Space。`Space` 保存於 `channels/{channelId}/spaces/{spaceId}`，由已驗簽的群組生命週期獨立建立／更新，候選索引及目前成員核驗獨立於 Repository Service。

Repository 的既有 ID、groupId 綁定、事項與討論保存位置不變。首版來源 Repository 的 groupId 必須等於擁有 Space 的 ID，拒絕跨 Space 引用；從 sourceRepositoryId 不得推導新增管理權或跳過當下來源授權。

## 操作契約

| 欄位 | Space 列表／詳情 | Project 列表／詳情／建立／更新 | ProjectItem 列表／加入／移除 |
| --- | --- | --- | --- |
| Actor | 既有 Session／Bearer 驗證的本人 | 同左 | 同左 |
| Resource／Action | channel 內 Space；List／Get | owner Space 內 Project；List／Get／Create／Update | Project 自有項目；List／Add／Remove |
| Context／Authority | 可信群組候選不代表授權；ACTIVE Space 與當下群組成員核驗 | Space 成員可讀／建立，Project 建立者管理內容；來源資格另外核驗 | Project 建立者管理收錄，來源事項仍依目前 Repository 讀取資格 |
| Policy／Decision | 非成員、停用、不存在不揭露；外部核驗失敗不當作允許 | 一個同 Space sourceRepositoryId；沿用內容限制、expectedVersion 與防重鍵；Project 無獨立完成狀態 | 同 Project 不重複，同事項可在不同 Project；項目 ID 與 Issue ID 分開，來源資格不允許則不揭露來源內容 |
| Result | 本人可開啟的 Space／基本資料，不授予 Issue 寫權 | Project 自有資料／版本；不寫 Repository、Issue 內容或狀態 | 僅改關聯及 Project 版本，來源事項完全不變；進度由可讀來源查得，不複製 Issue 狀態 |
| 重送／併發 | 讀取無副作用；外部成員資格不宣稱與資料庫原子 | 舊建立鍵及收據維持同結果／同鍵異內容衝突；更新過期版本拒絕，無變化不加版本 | expectedVersion 序列化收錄／移除，結果不明先讀取再決定；遷移鎖拒絕寫入，不另寫兩份 |
| Event／Audit | 保存必要群組事件排序，不記 token／profile | 保留原作者、建立／更新時間及版本追溯 | 關聯變動與 Project metadata 同交易，無外部通知 |

Project 自有資料的授權不得再要求 Repository 作為父容器。來源成員資格被撤銷或來源不可用時，來源 Issue 的內容、狀態及衍生進度不能由 Project 權限繞過；不得把未知統計回傳成零或以舊來源資料冒充成功。前端隱藏按鈕不能代替後端檢查。

## API、頁面與側欄

- Space 提供獨立 List／Get；Project 新入口為 `/api/v1/spaces/{spaceId}/projects` 及其專案／items 子資源。Create 明確指定 `sourceRepositoryId`，URL 的 Space 或 Project ID 不是授權憑證。
- 舊 `/api/v1/repositories/{repositoryId}/projects...` 保留相容 adapter，解析來源及 Space 後呼叫同一組用例，不維護第二份授權或寫入邏輯。原回應、ID、版本、防重及分頁契約依相容檢查保留。
- 舊Repository統計的projects欄位相容回報以該Repository為來源的專案數，依唯一保存位置讀取；這是唯讀來源關聯統計，不表示Repository擁有或可寫Project。
- 正式專案頁為 `/miniapp/projects`，URL 保存 Space／Project 選取；側欄「專案」與「儲存庫」同層。Project 頁透過 Space API 與本人 Session 管理資料，不需要 RepositoryProvider 作為 Project 生命週期。
- Repository 的專案頁籤僅作導向新入口的過渡連結。新增專案的來源選擇只列可讀來源，不能因此取得事項修改／審核權。切換 Space、Project、身份或登出時取消舊請求並隔離畫面資料，上一頁、重新整理與直接連結可回到同一目標。
- 原官方 shadcn primitives 及 sidebar-07 模板維持不變；業務導覽在應用層組合。Pages Router 維持靜態輸出，Go 靜態頁路由亦註冊新入口，不新增前端框架或依賴。

## 保存與遷移相容

正式 Project 保存於 `channels/{channelId}/spaces/{spaceId}/projects/{projectId}`，與 Repository 的舊保存位置分離。Project 保存 Space 及單一來源識別；items／changes 保留原 ID、版本、作者及時間。新增來源集合或跨 Space 功能須另定契約，首版不預建引擎。

既有 Repository ID 只用於遷移來源對照，不作 Space 身份。明確檢查 `channel → repository.groupId → SpaceID`，遇到無效身份、資料碰撞、異內容目標或孤立資料即拒絕，不以猜測覆蓋。建立收據移至 Space 下按來源隔離的 `projectCreationScopes/{sourceRepositoryId}/receipts`，保留原收據 ID 與payload雜湊，避免來源切換造成舊建立鍵重複。

過渡路由記錄為 `channels/{channelId}/projectStorage/{sourceRepositoryId}`；其 `spaceId`、`location`、`locked`、`migrationId` 只決定相容保存位置，不構成 Project 擁有權。所有相關寫入在交易內讀取此記錄；遷移時先鎖定寫入、複製及完整驗證，再切換唯一寫入位置，沒有雙寫。新來源沒有舊資料時使用 Space 保存；既有資料未遷移時以明確相容路由服務，不自動搬正式資料。

遷移工具預設 dry-run，實際apply／rollback需要明確指定，保留manifest及驗證結果。回復只在目標未有後續變更時允許，發現新寫入、不同鎖或hash不符即停止；不能為回復而丟失新資料。正式雲端apply、資料刪除與部署不包含在本次程式實作的授權內。

實際旗標、部署前提、writer確認及中斷處理依 [遷移操作指南](../operations/project-migration.md)。

待正式資料完成遷移、舊呼叫者與游標相容窗口結束後，才移除legacy保存resolver、路由marker及Repository轉接入口；必須另有使用證據與相容性決定，不因本文件自動刪除。

## 驗證邊界

驗收須包含 Space 與來源權限分離、非成員／停用拒絕、同 Space 來源限制、作者管理、項目唯讀來源、防重重送、版本衝突、舊ID／items／收據／游標相容、遷移dry-run無寫入、copy／validate／切換／回復及漂移拒絕。前端驗證側欄同層、直接URL、身份與選取切換、遲到回應、建立結果不明及靜態入口。

本機CLI、型別、測試、emulator及建置證據與真實 LINE／正式資料庫／手機／部署驗收分列；未實際執行的項目不標通過。工具及來源核對遵循 [業務操作](../architecture/business-operation.md) 與 [驗收規則](../agents/acceptance.md)。原本最小追蹤切片的歷史紀錄保留於 [planning](planning.md)，本頁為本次獨立 Project 的權威契約。

## 本次實際驗證（2026-11-13）

- 已維護Space service／Firestore及Project service／Firestore測試，另有migration CLI單元與emulator測試。隔離本機emulator的race驗證涵蓋無Repository的Space事件／候選、owner與來源撤權分離、非成員／作者／版本／同Space限制、canonical與legacy的同ID／收據重送、原Issue不變、marker鎖與部分copy不可見，以及rollback後新頁仍可讀legacy資料。
- 遷移實測dry-run零寫入、writer-ready門檻、完整items／changes／收據及時間版本保存、異內容碰撞／來源及目標漂移拒絕、無漂移回復與新寫入後不切回舊資料；舊Repository統計在移轉前後不重複計數。
- API能力檢查、生成同步及相容性驗證通過（61項操作、27個契約來源）；Go格式、lint、module verify、vet、完整既有race suite、build與漏洞檢查通過。漏洞檢查無應用可達項，仍有3筆未被應用呼叫的required-module項目，不稱所有依賴無漏洞。
- 前端型別及29個現有框架測試通過，包含新增6個Space／Project／身份與來源隔離、同層側欄、URL回復、原鍵重試及來源不可用情境；Next靜態build輸出新projects頁。官方UI primitives、use-mobile及sidebar-07模板與基準完全相同，未新增依賴。

LINE官方來源庫本次HEAD為 `e39b8daf919dc7a9396e44b5fc55a2aad918d932`，核對group-chats與其成員profile引用。Context7核對Next `16.4.5` Pages Router，並對照安裝版；Firestore官方Go client交易來源與本機 `1.29.2` 核對讀取先於寫入及callback可重試。Google Developer Knowledge／Google skills未提供可用入口，未宣稱其主機驗證或正式服務驗收。實際程式來源先用Graphify與Repomix定位／完整交叉核對，修改後重建並驗證圖與來源涵蓋；AST關係與安全掃描不取代執行測試。

本次僅對隔離emulator執行apply／rollback，沒有對正式資料庫執行遷移、清理或部署。正式環境須先確認writer升級與namespace、審閱dry-run manifest，依操作指南另行執行；有canonical新寫入的回復仍需額外copy-back方案，不在本工具能力內。
