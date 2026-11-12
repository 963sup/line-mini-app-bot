# Repository 共用資源編號基礎

狀態：僅內部基礎，尚未接入任何業務建立入口。使用者指定先做共用編號，不開放建立 Issue；公開與內部 OpenAPI、瀏覽器畫面及既有資料維持原樣。

## 範圍與識別

每個 Bot 管理的 Repository 共用一條資源編號序列，從 1 開始。例如 Issue #1、Discussion #2、Issue #3；不按資源類型另建計數器。另一個 Repository 從 1 開始。

`id` 是不透明的隨機識別碼，`number` 是資源在空間內的可見編號。資料保存整數，畫面接入時才加 `#`。單一號碼能解析為資源類型與 ID，但不是授權憑證。各業務仍擁有獨立的狀態、權限與資料。

本次未決定每一種資源是否需要編號。未來接入的業務若需要編號，一律使用同一入口；不預建 Issue、Discussion、Member、Label、Milestone 或 Project 的建立權限與生命週期。

## 保存結構

以下路徑均位於既有 `channels/{channelId}/repositories/{repositoryId}` 下：

| 路徑 | 責任 |
| --- | --- |
| `counters/resources` | 唯一計數器，schemaVersion 與 lastNumber |
| `resourceNumbers/{number}` | schemaVersion、repositoryId、resourceType、resourceId、number、createdAt |
| `resourceCreationReceipts/{hash}` | schemaVersion、payloadHash 與原始編號對照結果 |

資源類型使用穩定的小寫名稱；本基礎不接受客戶端直接選擇類型、操作者、ID 或號碼。收據鍵使用命名空間、可信操作者與防重鍵的 JSON tuple 做 SHA-256；channel 與 Repository 由保存路徑隔離。同一操作者在同一空間共用建立鍵範圍，跨類型重用鍵會回傳衝突。

初次成功寫入時才建立計數器，不由 Bot 建立 Repository 時預配。不回收、不重設、不因 Bot 離開或再加入而清零；不因刪除業務資料而刪除編號對照。上限為 JavaScript 可精確表示的最大整數 2^53−1。既有損壞計數器與下一號已存在時拒絕寫入，不覆蓋舊對照；不自動修復手動刪改的資料。

## 操作契約

| 欄位 | 建立帶編號的業務資源 |
| --- | --- |
| Actor / Resource / Action | 未來業務使用案例傳入的可信操作者；既有 Repository；CreateNumberedResource |
| Context / Authority | 呼叫者必須先驗證登入及該項業務建立權；GroupID 來自已授權的 Repository。本方法不是授權服務，不由讀取資格推導建立權 |
| Policy / Decision | 交易內讀取 Repository，確認啟用且群組綁定與已授權範圍一致；檢查防重收據、計數器與下一號對照。額外業務規則由 callback 在同一交易檢查 |
| Result | 同一交易提交業務資源、共用計數器、編號對照與防重收據；僅整個方法成功後可使用回傳結果 |
| 重送／併發 | 同操作者、鍵、類型與內容雜湊回原 ID／number／時間，不呼叫建立 callback、不增加計數。同鍵不同內容或類型回 ErrCreationConflict。不同請求透過共用計數器交易序列化 |
| Event / Audit | 對照保存類型、ID、號碼與時間；收據提供操作者／鍵範圍的防重證據。正式業務 Audit 由未來功能定義；不記錄 token 或原始防重鍵 |
| 失敗 | callback 或驗證拒絕會回滾本次交易；整體方法發生錯誤時回傳零值與錯誤。提交結果不明的網路錯誤必須沿用原鍵重試，不宣稱一定未提交 |

| 欄位 | 依編號解析 |
| --- | --- |
| Actor / Resource / Action | 未來已授權使用案例；指定 Repository 的編號；GetResourceNumber |
| Context / Authority | 此方法只做低層持久化查詢；對外使用前必須核驗容器及目標業務資源存取權，不能直接轉成無授權公開端點 |
| Policy / Decision | 檢查 ID、號碼範圍與對照資料完整性；無資料回 ErrNotFound，損壞或服務不可用則拒絕 |
| Result / 重送 / Audit | 回傳類型、ID、號碼與建立時間；無寫入、無額外事件，不表示使用者有權讀取目標內容 |

## 內部接入方式

- 核心型別與錯誤：[numbering.go](../../backend/internal/business/repository/numbering.go)，不依賴 Firestore／HTTP DTO。
- Firestore 實作：[numbering.go](../../backend/internal/business/repository/firestore/numbering.go)。使用既有 Store 的 client 與 channel，不另建服務或連線。
- `CreateNumberedResource(ctx, command, create)`：command 包含已核驗的 Repository／群組／操作者、資源類型、防重鍵及正規化且帶版本的業務內容雜湊。雜湊不包含伺服器產生的 ID／number／時間。
- `create(ctx, tx, reference)`：未來業務 Firestore adapter 依 reference 的 ID／number 建立自己的資料，所有讀寫使用同一 tx。可先讀取額外資料檢查業務規則，再寫入；不得提交交易、開新交易、呼叫外部服務或發送通知。
- 共通交易先完成全部共用讀取，再執行 callback，最後寫入計數器、對照與收據。callback 可能被重試，不得依靠單次執行或向外洩漏未提交結果。
- 隨機 ID 與微秒精度 UTC 時間在交易 callback 外產生，號碼在交易內決定。重試結果不更改原資源建立時間。
- `GetResourceNumber(ctx, repositoryID, number)`：只解析對照，未新增 API 路徑或修改生成型別。

群組成員核驗依既有外部機制完成，無法與 Firestore 組成原子快照；交易中的容器狀態檢查不擴大此保證。未來業務寫入權須另定，不把本基礎視為已授權。

## 查證與驗收範圍

Graphify 定位既有 Repository 與交易模式，Repomix 限定核對 Repository、契約雛形及既有補卡防重來源；指引與 Markdown 另讀原文。Context7 `/googleapis/google-cloud-go` 的官方 transaction.go 與本機 firestore v1.29.2 原始碼交叉確認：callback 可能重試、所有讀取須先於寫入，只有 RunTransaction 回傳 nil 才能確認提交。

Google Developer Knowledge 與 Google skills 未在此聊天提供可用入口，因此使用上述官方 Go client 來源；沒有執行 Google Cloud 資源操作。驗證結果於交付訊息列明，本機 emulator 與 race smoke 不取代正式環境驗收。沿用既有 CI，不新增正式測試框架或部署流程。

本次本機驗證：使用暫存 Go overlay 與隔離的 Firestore emulator 執行 race smoke，確認跨類型與跨空間計數、併發唯一號、同鍵併發重送、內容／類型衝突、業務寫入回滾、停用／群組綁定拒絕、再加入續號、損壞資料與整數上限及取消語義。驗證程式與 Java runtime 僅放 `/tmp`，未新增正式測試案例或修改專案工具版本。Graphify 更新有既有 Identity 方法名稱碰撞警告，該區不作為本次編號實作證據；文件圖譜的本機連結檢查通過。
