# 空間事項第一階段

狀態：已依本對話確認的產品規則完成本機實作，正式 LINE／Firestore 環境待驗收。

## 範圍與產品決策

2026-11-13 先以 Graphify CLI、Repomix CLI 及原始碼核對 [正式空間雛形](repository.md#正式空間-ui-基礎雛形)，未發現阻擋此讀取雛形的結構或方向問題，再進入事項切片。
使用者確認：核驗成員皆可讀取與建立事項；僅建立者可修改及變更自己事項的狀態。首版採待處理、進行中、已完成，不含指派或審核。

每個事項歸屬既有 Repository；不新增空間、人類管理員、刪除、附件、通知、留言或專案功能。
讀取資格與寫入政策分開檢查，不影響出勤、補卡或原群組摘要授權。事項完成不代表施工驗收或業務核准。

## 操作契約

| 欄位 | 列表／詳情 | 建立 | 修改／生命週期 |
| --- | --- | --- | --- |
| Actor | 後端既有 Session／Bearer 驗證身分 | 同左，建立者由伺服器取得 | 同左，不接受客戶端宣告角色 |
| Resource / Action | 指定 Repository 的 Issue；List / Get | 指定 Repository；CreateIssue | 指定事項；UpdateIssue（內容或明確的目標狀態） |
| Context | 啟用空間、當下成員資格、搜尋與游標 | 啟用空間、正規化輸入、原建立鍵 | 啟用空間、目前事項、預期版本 |
| Authority / Policy | 該空間目前核驗成員可讀取 | 該空間目前核驗成員可建立，非成員拒絕 | 必須仍是成員且為建立者；版本及合法轉換在交易內核對 |
| Decision | 來源空間及資格符合才回傳；不存在或無權空間統一 404 | 權限、輸入及防重條件皆符合才建立 | 權限、版本與狀態皆符合才更新；其他成員不能修改 |
| Result | 回傳符合可見範圍的事項，不變更資料 | 建立 open 事項；回傳伺服器 ID、共用 number 與保存快照 | 回傳最新版本；拒絕不改變事項；相同版本及相同內容不增加版本 |
| 重送／併發 | 游標綁定 channel、空間、本人及篩選 | 沿用 [共用編號](repository-numbering.md) 的原子建立、防重收據與衝突規則 | 過期版本回 409；結果不明時先讀取，不能直接把原更新當另一項意圖重送 |
| Event / Audit | 一般讀取不額外保存個資 | 事項保存建立者及建立時間 | 每次有效更新保存操作者、版本、時間與前後狀態；無外部通知 |

狀態值為 `open`、`in_progress`、`completed`。唯一跨狀態轉換為 `open → in_progress → completed → open`；同狀態可修改內容。不可跳過進行中直接完成。
標題去除首尾空白後須有 1–120 個 Unicode 字元；內容去除首尾空白後最多 8000 個字元。HTTP JSON 上限 65536 bytes，不接受未知欄位。識別碼、number、操作者與版本來源不可由建立請求指定。

## API、查詢及資料

契約唯一來源為 [issues.yaml](../../api/contract/issue/issues.yaml)，由既有 OpenAPI 入口組合及生成 Go／TS 型別。

- `GET /api/v1/repositories/{repositoryId}/issues`：可選 cursor、state、query；依共用 number 遞減。
- `POST /api/v1/repositories/{repositoryId}/issues`：必填 Idempotency-Key、title、body。
- `GET /api/v1/repositories/{repositoryId}/issues/{id}`：查詢詳情。
- `PATCH /api/v1/repositories/{repositoryId}/issues/{id}`：完整提交 title、body、state、expectedVersion；僅建立者。

一次最多掃描二十筆候選，搜尋以正規化標題、不區分大小寫的子字串比對，再套用狀態篩選。可能回傳空頁但仍有 nextCursor，前端必須提供繼續載入；不將單頁筆數作總數，不新增複合索引。游標無效或換空間／身分／條件後沿用會被拒絕。
建立鍵與同空間同操作者的其他資源共用範圍；相同鍵及內容回同一 ID／number 的目前快照，內容不同回 409。建立雜湊使用 adapter 明確編碼的 v1 正規化 title／body，不雜湊 HTTP DTO。

資料保存於既有 `channels/{channelId}/repositories/{repositoryId}/issues/{id}`；有效更新的追溯資料位於其 `changes/{version}`。建立與空間共用編號同一交易提交；更新交易重讀空間、事項、建立者及版本，再提交事項與追溯資料。Bot 停用不刪除事項，但拒絕新的讀寫；再啟用沿用資料。
LINE 成員核驗與 Firestore 無法組成原子快照，沿用既有外部資格觀察限制；寫入交易另檢查容器狀態及群組綁定。

失敗語義：無效輸入 400、未登入 401、非建立者／CSRF 403、無權空間或不存在事項 404、過期版本或建立鍵衝突 409、非 JSON 415、服務未配置／不可用／提交結果不明 503。錯誤不回傳 token、輸入內容或內部 SDK 原因。

## 實作歸屬與前端

- [核心](../../backend/internal/business/issue/issue.go)：用例、建立資格、建立者權限、內容正規化與狀態轉換；不依賴 HTTP／Firestore DTO。
- [HTTP](../../backend/internal/business/issue/http/routes.go)：生成路由與 DTO，沿用既有身份／CSRF，認證後驗證契約，再進用例。
- [Firestore](../../backend/internal/business/issue/firestore/store.go)：文件轉換、查詢、建立防重與交易；由 server 注入既有 Repository client 及編號介面，不另建連線。
- [事項畫面](../../frontend/miniapp/src/business/issue/View.tsx)：由正式空間入口組合，透過既有 Session client 載入；不將 API 或規則放進 Shell／shared/ui。

列表、詳情、寫入各有請求取消及世代檢查；登出、重新驗證或空間切換清除舊內容。新增連點由同步鎖阻擋；結果不明時鎖定原內容並沿用建立鍵重試。修改失敗後先重新讀取，不覆蓋新版本。建立鍵僅保留於本次畫面，重新開頁或離開事項前應先確認不明結果，不能宣稱跨重新整理防重。

## 查核來源與驗證邊界

Graphify 定位共用 navigation、Repository Context 及業務關係；同名 View 與動態 JSX 另以原始碼／Repomix 交叉核對，不以圖譜無路徑推論無實作。
第一階段六個檔案與先前驗證包相符；沿用當時的型別檢查、23 個前端測試、靜態建置及七個入口切換證據，事項變更另執行受影響檢查。
Next.js 架構沿用已查閱的 Context7 `/vercel/next.js/v16.4.5` 及安裝版 Pages Router 文件；UI 依固定 shadcn CLI 的官方範例組合，未改 primitives。
Google skills／Developer Knowledge MCP 未提供可用入口；改查 Context7 `/googleapis/google-cloud-go` 官方 transaction.go，並核對本機 firestore v1.29.2 的交易重試、先讀後寫與游標行為。未操作正式 Google Cloud 資源。
實際檢查結果於交付訊息列出；本機檢查不取代真實 LINE 裝置、外部身分及正式資料庫驗收。

本次本機驗證：API 能力／生成碼同步／向後相容、Go 格式／lint／vet／完整既有 race suite／build、前端型別／23 個既有測試／靜態輸出、設定及文件路徑檢查通過。govulncheck 未發現應用可達漏洞，另報三筆未被程式呼叫的 required-module 項目。
隔離 Firestore emulator 的暫存 Go overlay race smoke 確認共用編號、併發建立防重、建立鍵衝突、目前成員與建立者規則、非法狀態拒絕、版本衝突及併發更新只允許一筆成功、追溯紀錄、空頁續查、游標範圍、跨空間隔離、生成 HTTP 驗證、停用容器交易拒絕及取消語義。
暫存 React smoke 確認正式空間事項入口、建立連點鎖、結果不明重試沿用鍵、三狀態操作、版本衝突後重新讀取、他人事項不顯示寫入操作及登出後拒收遲到回應。smoke 與 Java runtime 僅置於 `/tmp`，未新增正式測試框架或部署設定。

進入討論階段前再次查核，修正 Firestore 未保留未知 SDK 原因的缺口：現在以私有 storageFailure 保存原因，Unwrap 只公開 ErrUnavailable，維持原 HTTP 失敗語義；受影響套件與分層檢查、後續完整既有 race suite 均通過。

標籤分類透過既有更新的可選 labelIds 接入，只有事項作者可套用或移除；省略欄位保留、空陣列清除、同空間及版本規則依 [標籤引用切片](labels.md#標籤引用切片)。既有建立防重雜湊不變。
