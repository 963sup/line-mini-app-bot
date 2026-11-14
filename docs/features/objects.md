# 儲存庫工程項目 Objects

狀態：最小切片已完成本機實作與回歸驗證；正式 LINE／Firestore／手機環境待驗收。每個來源文件項次是一筆 Object，quantity 代表該項數量，不自動拆成單件資產。

## 範圍與欄位

儲存庫內提供「工程項目」列表、詳情、手動新增與編輯。類型為人工確認文字，可沿原文件分類群組；核心不負責分類目錄、AI 模型、檔案解析、附件、價格／稅／付款／驗收、庫存、階層、問題關聯或刪除。文字型 PDF 入口由獨立匯入流程提供，仍需人工審核。

| 欄位 | 規則 |
| --- | --- |
| name、description | 名稱 1–120 字元、完整說明最多 8000 字元；首尾空白移除，不靜默截斷 |
| type、unit | 必填人工分類 1–80 字元、原單位 1–32 字元；不拿共同料號當類型，不做單位換算 |
| quantity | 精確正十進位字串，最多 18 位整數及 6 位小數；不用浮點數。拒絕零、負數、指數、千分位，整數前導零只允許單一 0，小數尾零正規化後保存 |
| status | not_started（未開始）／in_progress（進行中）／completed（完成）；建立固定未開始，之後三態可互轉。只是人工追蹤，完成不代表驗收或核准 |
| sourceDocument、sourceRevision、sourceItemNo、sourcePages、materialCode | 可選來源追溯文字，各上限 200／80／80／80／120 字元；不儲存原件、個資或價格。來源文件與 revision 分開，頁碼可記跨頁範圍 |
| id、repositoryId、creatorId、version、createdAt、updatedAt | 後端可信來源，建立不可自行指定；不提供額外 #number |

來源欄位可獨立留白，只是追溯資料，不承諾 sourceDocument＋revision＋itemNo 唯一。同來源換建立鍵仍可新增另一筆；使用者須人工核對。新版 PDF 不會自動比對、合併或覆寫既有工程項目與狀態。來源項次不作為資料庫 ID。

## 操作契約

| 欄位 | List／Get | Create | Update |
| --- | --- | --- | --- |
| Actor／Resource／Action | Session／Bearer 可信使用者；指定 Repository 的 Object；讀取 | 同左；建立 Object | 同左；編輯既有 Object |
| Context／Authority | 啟用儲存庫、當下群組成員資格 | 同左；明確允許目前成員建立 | 目前成員且為建立者；目前資料與 expectedVersion |
| Policy／Decision | 容器與成員核驗通過才回傳；無權／不存在不揭露 | 驗證內容、建立鍵及容器；身分不能由客戶端宣告 | 交易內核對容器、作者、版本及輸入；拒絕不改資料 |
| Result | 列表／詳情，無副作用 | 伺服器產生 ID、初始狀態、version=1 與時間 | 有效變更遞增版本，回傳保存快照；相同內容不製造變更 |
| 重送／併發 | 游標綁定授權範圍及本人，分頁不可當總數 | Idempotency-Key 同本人／儲存庫／內容回原資源，不新增；同 key 不同內容 409。未知結果保留原 key 重試 | expectedVersion 過期 409，先重讀再決定，不能無條件覆寫 |
| Event／Audit | 不額外記錄讀取個資 | 保存建立者／時間與防重收據 | 保存操作者、版本、時間與變更；不含 token，無通知副作用 |

群組成員核驗屬當下外部觀察，不能和 Firestore 組成原子快照；交易仍核對容器啟用及群組綁定。讀取資格不自動授予修改他人資源的權限。身分或來源服務不可用時拒絕，不以快取或 UI 按鈕取代授權。

## API 與保存

權威來源為 [Objects 契約](../../api/contract/object/objects.yaml)，由既有 OpenAPI 入口生成 Go／TS 型別：

- GET／POST `/api/v1/repositories/{repositoryId}/objects`；列表依不透明文件 ID 遞增分頁，不宣稱為建立時間或來源項次順序。
- GET／PATCH `/api/v1/repositories/{repositoryId}/objects/{id}`。
- 建立只收內容與可選來源欄位；狀態由後端設定。PATCH 收完整內容、status、expectedVersion；可選來源欄位省略時保留現值，明確空字串才清除；不做文件版本自動合併。

資料位於 `channels/{channelId}/repositories/{repositoryId}/objects/{id}`；不改 Issue、共用編號計數器或既有資料。建立資源與防重收據同交易提交；交易 callback 不呼叫外部服務或發通知。

結構化候選的獨立 [匯入草稿基礎](object-imports.md) 另外協調驗證／人工確認，仍只呼叫本 Objects 建立使用案例；不讓核心依賴匯入或 AI。

## 畫面與驗收

儲存庫工程項目頁籤使用既有元件與 typed client。只有作者顯示編輯；切換身份／儲存庫隔離舊回應，失敗保留可重試意圖。未知建立結果保留原 key 與內容，不以另一鍵重複提交。

最小驗收包含精確數量／非法輸入、預設及三態轉換、未登入／非成員／停用／跨儲存庫拒絕、非作者修改拒絕、版本衝突、防重重送與內容衝突、未知結果重試、前端載入／失敗／切換身份及手機表單。使用脫敏測資；真實商業 PDF 不進儲存庫或公開 fixtures。

實作位置：後端 [object](../../backend/internal/business/object/README.md)、前端 [object](../../frontend/miniapp/src/business/object/README.md)；正式 LINE／Firestore／手機整合另行驗收，測試替身或 build 不代表正式環境通過。本輪不自動匯入來源文件的 54 項。

[功能索引](README.md) · [文件索引](../README.md) · [操作模型](../architecture/business-operation.md)
