# 空間標籤目錄

狀態：依使用者「先用最基礎的實現」完成目錄列出與建立及本機隔離驗證；正式環境待驗收。

## 前置查核與最小範圍

先以 Graphify CLI、Repomix CLI 核對討論／留言原始碼，與上一階段驗證包一致，沿用已有契約、分層、交易及 UI 驗證證據。
查核發現空間文件仍稱共用編號未接入業務，已修正為事項與討論已接入、留言不占號；程式未發現阻擋此切片的結構或方向問題。

沿用現有成員核驗模式，核驗成員可列出及建立空間標籤。目錄只有名稱，不新增改名、刪除、自訂顏色、角色或編號；本次以既有事項／討論的作者編輯權限接入標籤引用，不擴大可編輯對象。

## 操作契約

| 欄位 | 列出目錄 | 建立標籤 |
| --- | --- | --- |
| Actor / Resource / Action | 既有 Session／Bearer 驗證本人；指定 Repository 的 Label；List | 同左；CreateLabel，作者由伺服器取得 |
| Context / Authority | 啟用空間、當下群組成員資格、本人範圍游標 | 同左；正規化名稱與原建立鍵 |
| Policy / Decision | 只回傳該空間目前核驗成員可讀取的標籤 | 核驗成員可建立；同一空間名稱比較值必須唯一，交易內確認容器狀態與群組綁定 |
| Result | 最多二十筆及 nextCursor，不變更資料 | 保存不透明 ID、名稱、建立時間與可信作者；不消耗共用 number |
| 重送／併發 | 游標綁定 channel、空間與本人，依 ID 穩定遞增 | 同空間、本人及建立鍵重送相同名稱回同一 ID；同鍵不同名稱或同空間同名回 409，不重複建立 |
| Event / Audit | 一般讀取不額外保存個資 | 作者與建立時間保存於標籤，收據提供防重證據；無外部通知 |

名稱去除首尾空白後 1–40 個 Unicode 字元，不得含控制字元；名稱比較值使用 strings.ToLower，不宣稱 Unicode NFC 或完整 case-fold 等價。
名稱唯一性限於同一 Repository，不因另一個空間有同名標籤而拒絕。JSON 上限 1024 bytes，不接受未知欄位；ID、時間及作者不能由客戶端指定。首版建立後不提供修改或刪除。

## API、資料與結構

契約來源為 `api/contract/repository/labels.yaml`，由既有 OpenAPI 入口生成 Go／TS 型別，不變更相容性 baseline。

- `GET /api/v1/repositories/{repositoryId}/labels?cursor=...`：列出可讀目錄，每頁二十筆。
- `POST /api/v1/repositories/{repositoryId}/labels`：body 只有 name，必填 Idempotency-Key；同一次未確認建立重試沿用原鍵及內容。

資料位於既有 `channels/{channelId}/repositories/{repositoryId}` 下的 `labels/{id}`；`labelNames/{hash}` 保存名稱唯一綁定，`labelCreationReceipts/{hash}` 保存本空間、本人及建立鍵的收據。建立以同一交易保存標籤、名稱綁定與收據，所有讀取先於寫入。
Bot 停用保留資料、拒絕新讀寫；再次啟用沿用標籤。LINE 成員核驗是外部資格觀察，不宣稱與 Firestore 原子一致。錯誤只公開領域分類，SDK 原因保留在 adapter 私有值。

核心位於 [repository/labels](../../backend/internal/business/repository/labels/labels.go)，HTTP 在其 `http` 子目錄；[Firestore](../../backend/internal/business/repository/firestore/labels.go) 使用既有 Repository Store，不新增 client。[前端](../../frontend/miniapp/src/business/repository/labels/View.tsx) 由正式空間「標籤」入口組合現有 UI，不改 primitives 或 Shell 業務責任。
失敗語義：400 輸入／游標、401 身分、403 CSRF、404 無權或不存在空間、409 同名／建立鍵衝突、415 非 JSON、503 未配置／不可用／結果不明。
前端分開取消讀寫請求，離開空間或身分變更時丟棄舊回應；同步鎖阻擋連點，建立結果不明時鎖定原名稱並沿用鍵重試。鍵僅保留本次畫面，離頁或重新整理前應先確認結果。

## 完成後評估

目錄本身不授予事項／討論寫入權。本次已接入下述標籤引用切片，保留各自作者與版本規則；基礎雛形完成界線見 [空間雛形](repository-prototype.md)。
在擴大功能前，仍需實際 LINE 身分、群組成員與正式資料庫的整合驗收，本機模擬器不取代這些證據。里程碑與專案管理排在標籤引用後；里程碑完成規則、專案歸屬與項目狀態應先定義，不由預覽百分比或事項狀態猜測。

| 優先順序 | 後續工作 | 評估與邊界 |
| --- | --- | --- |
| 1 | 實際整合驗收 | 目前只完成本機／隔離驗證，正式 LINE 裝置與資料庫證據仍缺；部署前先補齊，不因目錄完成就宣稱上線 |
| 已接入 | 標籤套用到事項／討論 | 使用既有更新 API，保留作者及版本規則，仍不加入標籤改名／刪除 |
| 3 | 正式總覽摘要 | 現況只有空間資料與功能入口；可先組合既有事項／討論查詢，不放示範統計或以單頁筆數作總數 |
| 4 | 里程碑、專案與稽核畫面 | 延後；先定義完成條件、項目來源、狀態歸屬及可見權限；既有更新 metadata 不等於完整公開稽核系統 |

## 查證與驗證

沿用本對話已查閱的 Context7 官方 Google Go client 交易契約及本機 firestore v1.29.2：所有讀取先於寫入，callback 可能重試，只有交易回傳 nil 才確認提交。Google skills／Developer Knowledge MCP 未提供可用入口；本次未操作正式資源。
錯誤邊界沿用官方 errors.Is／Unwrap 語意及既有 adapter 模式，Go 1.27.4；UI 沿用固定 shadcn 官方範例與 Next.js 16.4.5 Pages Router 架構，不新增套件或改寫 primitives。Graphify 與 Repomix 用於定位及原文交叉核對，不取代行為驗證。

API 能力／生成同步／向後相容、Go 格式／lint／vet／完整既有 race suite／build、前端型別／23 個既有測試／靜態輸出、設定與文件連結檢查通過。govulncheck 未發現應用可達漏洞，另報三筆未被程式呼叫的 required-module 項目。
隔離 Firestore emulator 的暫存 overlay race smoke 確認正規化同名拒絕、不同空間同名允許、同鍵重送、同鍵異名拒絕、併發名稱唯一及重送只建立一次、分頁／游標範圍、成員資格、不占共用編號、HTTP 契約／CSRF／大小限制、無修改或刪除端點、停用交易拒絕及取消語義。
暫存 React smoke 確認正式目錄入口、連點鎖、原鍵重試、查詢失敗時清空舊目錄而不顯示成功空態、切換身分後拒收遲到標籤。驗證程式與 Java runtime 僅放 `/tmp`，未新增正式測試框架；不代表正式環境或真實 LINE 裝置已驗收。

## 標籤引用切片

| 欄位 | 套用／移除事項或討論的標籤 |
| --- | --- |
| Actor / Resource / Action | 現有 Session／Bearer 驗證本人；指定空間內的 Issue／Discussion；沿用 UpdateIssue / UpdateDiscussion |
| Context / Authority | 啟用空間、當下成員資格、目前資源、預期版本；只有該資源作者可編輯，不從標籤建立者推導內容修改權 |
| Policy / Decision | 每筆最多十個不同的同空間標籤 ID；在更新交易內讀取並驗證引用，再與作者／版本規則一起決定是否寫入 |
| Result | 更新資源自己的 labelIds 與版本，原標籤及共用編號不變；非法、他人、外空間或過期版本均不寫入 |
| 重送／併發 | labelIds 省略保留既有值，[] 清除；null、重複、格式錯誤、超過十個或不可用 ID 拒絕。ID 集合排序後比較，同集合換順序不增加版本；過期版本仍 409 |
| Event / Audit | 沿用各資源有效更新的操作者／版本／時間 metadata，不新增通知或公開稽核端點 |

透過既有 PATCH 事項／討論端點的可選 labelIds 欄位實作，不新增端點；標籤建立或讀取本身不改變關聯。建立事項／討論後，在詳情選取並儲存分類標籤。留言不套用標籤。
既有未含 labelIds 的資料讀作空集合，HTTP 新回應提供陣列（可選欄位，不回 null），舊更新請求省略欄位時不清空關聯。既有建立收據 v1 雜湊與 ID／number 分配完全不變，不掃描或回填歷史資料。
核心只共用 labels.NormalizeIDs 這個實際跨業務需要的值驗證；各領域保留自己的用例及版本規則。Firestore adapters 透過小型 LabelValidator 介面，在同一交易的寫入前驗證引用，由 server 注入既有 Repository Store，不建立共通操作引擎或新 client。
前端共用 business/repository/labels/Picker，透過既有目錄 API 分頁讀取名稱；載入或失敗不清除目前選取，讀取更多頁不觸發表單送出。非作者只能查看，作者選取／移除後沿用資源儲存與衝突處理。
