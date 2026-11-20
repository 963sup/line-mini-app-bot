# 空間標籤目錄

<a id="空間標籤目錄"></a>

```mermaid
flowchart LR
  h["空間標籤目錄"]
  h --- n0["狀態：依使用者「先用最基礎的實現」完成目錄列<br/>出與建立及本機隔離驗證；正式環境待驗收。"]
```

## 分類協作與改名

分類操作只更新同庫的標籤引用，以及 Issue 可選的里程碑。Go 重新核驗操作者資格、資源歸屬與版本。只有分類權的操作者若同時修改正文或問題狀態，整筆拒絕。Discussion 沿用同一套標籤規則。

授權改名時保留 Label ID；新名稱仍需在同庫唯一，名稱綁定與標籤在同一交易更新。以原建立鍵重送原請求，應取回同一 ID 與目前名稱。改名不需要逐筆更新 Issue 或 Discussion。群組 Repository 的所有目前成員均可分類及改名；個人 Repository 僅限本人。改名須提交讀到的原名稱，名稱過期回 409。

<a id="前置查核與最小範圍"></a>

```mermaid
flowchart LR
  h["首版歷史範圍"]
  h --- n0["先以 Graphify CLI、Repomix CLI 核對討論／留<br/>言原始碼，與上一階段驗證包一致，沿用已有契約<br/>、分層、交易及 UI 驗證證據。 查核發現空間文<br/>件仍稱共用編號未接入業務，已修正為事項與討論<br/>已接入、留言不占號；程式未發現阻擋此切片的結<br/>構或方向問題。"]
  h --- n1["沿用現有成員核驗模式，核驗成員可列出及建立空<br/>間標籤。目錄只有名稱，不新增改名、刪除、自訂<br/>顏色、角色或編號；本次以既有事項／討論的作者<br/>編輯權限接入標籤引用，不擴大可編輯對象。"]
```

<a id="操作契約"></a>

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["欄位：Actor / Resource / Action<br/>列出目錄：既有 Session／Bearer 驗證本人；指<br/>定 Repository 的 Label；List<br/>建立標籤：同左；CreateLabel，作者由伺服器取<br/>得"]
  h --- n1["欄位：Context / Authority<br/>列出目錄：啟用空間、當下群組成員資格、本人範<br/>圍游標<br/>建立標籤：同左；正規化名稱與原建立鍵"]
  h --- n2["欄位：Policy / Decision<br/>列出目錄：只回傳該空間目前核驗成員可讀取的標<br/>籤<br/>建立標籤：核驗成員可建立；同一空間名稱比較值<br/>必須唯一，交易內確認容器狀態與群組綁定"]
  h --- n3["欄位：Result<br/>列出目錄：最多二十筆及 nextCursor，不變更資<br/>料<br/>建立標籤：保存不透明 ID、名稱、建立時間與可<br/>信作者；不消耗共用 number"]
  h --- n4["欄位：重送／併發<br/>列出目錄：游標綁定 channel、空間與本人，依 <br/>ID 穩定遞增<br/>建立標籤：同空間、本人及建立鍵重送相同名稱回<br/>同一 ID；同鍵不同名稱或同空間同名回 409，不<br/>重複建立"]
  h --- n5["欄位：Event / Audit<br/>列出目錄：一般讀取不額外保存個資<br/>建立標籤：作者與建立時間保存於標籤，收據提供<br/>防重證據；無外部通知"]
```

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["名稱去除首尾空白後 1–40 個 Unicode 字元，不<br/>得含控制字元；名稱比較值使用 strings.ToLower<br/>，不宣稱 Unicode NFC 或完整 case-fold 等價。<br/> 名稱唯一性限於同一 Repository，不因另一個空<br/>間有同名標籤而拒絕。JSON 上限 1024 bytes，不<br/>接受未知欄位；ID、時間及作者不能由客戶端指定<br/>。首版建立後不提供修改或刪除。"]
```

<a id="api資料與結構"></a>

```mermaid
flowchart LR
  h["API、資料與結構"]
  h --- n0["契約來源為 <br/>api/contract/repository/labels.yaml，由既有 <br/>OpenAPI 入口生成 Go／TS 型別，不變更相容性 <br/>baseline。"]
  h --- n1["GET /api/v1/repositories/{repositoryId}<br/>/labels?cursor=...：列出可讀目錄，每頁二十筆<br/>。"]
  h --- n2["POST /api/v1/repositories/{repositoryId}<br/>/labels：body 只有 name，必填 <br/>Idempotency-Key；同一次未確認建立重試沿用原<br/>鍵及內容。"]
  h --- n3["資料位於既有 channels/{channelId}<br/>/repositories/{repositoryId} 下的 labels/{id<br/>}；labelNames/{hash} 保存名稱唯一綁定，<br/>labelCreationReceipts/{hash} 保存本空間、本<br/>人及建立鍵的收據。建立以同一交易保存標籤、名<br/>稱綁定與收據，所有讀取先於寫入。 Bot 停用保<br/>留資料、拒絕新讀寫；再次啟用沿用標籤。LINE <br/>成員核驗是外部資格觀察，不宣稱與 Firestore <br/>原子一致。錯誤只公開領域分類，SDK 原因保留在<br/> adapter 私有值。"]
  h --- n4["核心位於 repository/labels，HTTP 在其 http <br/>子目錄；Firestore 使用既有 Repository Store<br/>，不新增 client。前端 由正式空間「標籤」入口<br/>組合現有 UI，不改 primitives 或 Shell 業務責<br/>任。 失敗語義：400 輸入／游標、401 身分、403<br/> CSRF、404 無權或不存在空間、409 同名／建立<br/>鍵衝突、415 非 JSON、503 未配置／不可用／結<br/>果不明。 前端分開取消讀寫請求，離開空間或身<br/>分變更時丟棄舊回應；同步鎖阻擋連點，建立結果<br/>不明時鎖定原名稱並沿用鍵重試。鍵僅保留本次畫<br/>面，離頁或重新整理前應先確認結果。"]
  n4 --- n4l0["repository/labels"]
  click n4l0 "../../../backend/internal/business/repository/domain/entity/labels/labels.go"
  n4 --- n4l1["Firestore"]
  click n4l1 "../../../backend/internal/business/repository/adapter/firestore/labels.go"
  n4 --- n4l2["前端"]
  click n4l2 "../../../frontend/miniapp/src/business/repository/labels/View.tsx"
```

<a id="標籤目錄完成當時的評估"></a>

```mermaid
flowchart LR
  h["標籤目錄完成當時的評估"]
  h --- n0["下表記錄當時的排序；里程碑現況見里程碑規格。<br/>目錄本身不授予事項／討論寫入權。本次已接入下<br/>述標籤引用切片，保留各自作者與版本規則；基礎<br/>雛形完成界線見 空間雛形。 在擴大功能前，仍需<br/>實際 LINE 身分、群組成員與正式資料庫的整合驗<br/>收，本機模擬器不取代這些證據。里程碑與專案管<br/>理排在標籤引用後；里程碑完成規則、專案歸屬與<br/>項目狀態應先定義，不由預覽百分比或事項狀態猜<br/>測。"]
  n0 --- n0l0["里程碑規格"]
  click n0l0 "milestones.md"
  n0 --- n0l1["空間雛形"]
  click n0l1 "acceptance.md"
```

```mermaid
flowchart LR
  h["標籤目錄完成當時的評估"]
  h --- n0["優先順序：1<br/>後續工作：實際整合驗收<br/>評估與邊界：目前只完成本機／隔離驗證，正式 <br/>LINE 裝置與資料庫證據仍缺；部署前先補齊，不<br/>因目錄完成就宣稱上線"]
  h --- n1["優先順序：已接入<br/>後續工作：標籤套用到事項／討論<br/>評估與邊界：使用既有更新 API，保留作者及版本<br/>規則，仍不加入標籤改名／刪除"]
  h --- n2["優先順序：3<br/>後續工作：正式總覽摘要<br/>評估與邊界：現況只有空間資料與功能入口；可先<br/>組合既有事項／討論查詢，不放示範統計或以單頁<br/>筆數作總數"]
  h --- n3["優先順序：4<br/>後續工作：里程碑、專案與稽核畫面<br/>評估與邊界：延後；先定義完成條件、項目來源、<br/>狀態歸屬及可見權限；既有更新 metadata 不等於<br/>完整公開稽核系統"]
```

<a id="查證與驗證"></a>

```mermaid
flowchart LR
  h["查證與驗證"]
  h --- n0["沿用本對話已查閱的 Context7 官方 Google Go <br/>client 交易契約及本機 firestore v1.29.2：所<br/>有讀取先於寫入，callback 可能重試，只有交易<br/>回傳 nil 才確認提交。Google skills／<br/>Developer Knowledge MCP 未提供可用入口；本次<br/>未操作正式資源。 錯誤邊界沿用官方 errors.Is<br/>／Unwrap 語意及既有 adapter 模式，Go 1.27.4<br/>；UI 沿用固定 shadcn 官方範例與 Next.js <br/>16.4.5 Pages Router 架構，不新增套件或改寫 <br/>primitives。Graphify 與 Repomix 用於定位及原<br/>文交叉核對，不取代行為驗證。"]
  h --- n1["API 能力／生成同步／向後相容、Go 格式／lint<br/>／vet／完整既有 race suite／build、前端型別<br/>／23 個既有測試／靜態輸出、設定與文件連結檢<br/>查通過。govulncheck 未發現應用可達漏洞，另報<br/>三筆未被程式呼叫的 required-module 項目。 隔<br/>離 Firestore emulator 的暫存 overlay race <br/>smoke 確認正規化同名拒絕、不同空間同名允許、<br/>同鍵重送、同鍵異名拒絕、併發名稱唯一及重送只<br/>建立一次、分頁／游標範圍、成員資格、不占共用<br/>編號、HTTP 契約／CSRF／大小限制、無修改或刪<br/>除端點、停用交易拒絕及取消語義。 暫存 React <br/>smoke 確認正式目錄入口、連點鎖、原鍵重試、查<br/>詢失敗時清空舊目錄而不顯示成功空態、切換身分<br/>後拒收遲到標籤。驗證程式與 Java runtime 僅放<br/> /tmp，未新增正式測試框架；不代表正式環境或<br/>真實 LINE 裝置已驗收。"]
```

<a id="標籤引用切片"></a>

## 標籤引用切片

現行實作

[問題與討論的分類更新，以及里程碑本身的狀態轉換（Mermaid 原始檔）](diagrams/resource-classification.mermaid)

```mermaid
flowchart LR
  h["標籤引用切片"]
  h --- n0["Mermaid 原始檔。問題與討論沿用既有 PATCH，留<br/>言不套用標籤。標籤建立者沒有因此取得修改他人<br/>內容的權限。過期版本回 409；任何引用驗證失敗<br/>都不提交更新。成功只改該資源與既有更新紀錄（<br/>操作者、版本、時間），標籤與共用編號不變，也<br/>不產生通知或公開稽核端點。"]
  click n0 "diagrams/resource-classification.mermaid"
  h --- n1["既有未含 labelIds 的資料讀作空集合，HTTP 新<br/>回應提供陣列（可選欄位，不回 null），舊更新<br/>請求省略欄位時不清空關聯。既有建立收據 v1 雜<br/>湊與 ID／number 分配完全不變，不掃描或回填歷<br/>史資料。 核心只共用 labels.NormalizeIDs 這個<br/>實際跨業務需要的值驗證；各領域保留自己的用例<br/>及版本規則。Firestore adapters 透過小型 <br/>LabelValidator 介面，在同一交易的寫入前驗證<br/>引用，由 bootstrap 注入既有 Repository Store，<br/>不建立共通操作引擎或新 client。 前端共用 <br/>business/repository/labels/Picker，透過既有<br/>目錄 API 分頁讀取名稱；載入或失敗不清除目前<br/>選取，讀取更多頁不觸發表單送出。目前成員選取或移除分類後沿用資源版本與衝突處理。"]
```
