# 個人與群組-repository

**已有實作** · [領域設計](#領域設計) · [核心流程](#decision-flow) · [判定實作](../../../backend/internal/business/repository/domain/entity/repository.go) · [程式目錄](../../../backend/internal/business/repository/README.md) · [維護指引](../../../backend/internal/business/repository/AGENTS.md)

## 領域設計

**子域定位**　工程資源管理的核心候選。

**通用語言**　儲存庫具有識別碼、擁有者及啟用狀態；空間是庫內工作位置，與跨庫專案分開管理。

**上下文與契約**　Issue、Discussion、Object 等能力使用 Repository 的範圍與狀態；各自在交易中以最新儲存庫資料確認能否寫入。

**驗收案例**　即使庫 ID 相符，只要擁有者已變更或狀態不再是 ACTIVE，就不能沿用舊授權完成受限操作。

[對照實作或測試](../../../backend/internal/business/repository/domain/entity/repository.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

## 用途與範圍

保存 Repository 及其空間、標籤、事項、討論與其他子能力的規格。

## 架構與責任邊界

Repository 範圍及功能授權在此說明，各子能力細節由對應文件維護。

## 閱讀與修改入口

先查閱[spaces.md](spaces.md)、[議題規格](../issue/design.md)、[組織與團隊歸屬契約](../../architecture/contracts/organization-team-ownership.md)。

## 相關規格

[Repository Insights 與 Pulse](insights.md) · [工程知識庫 Wiki](wiki.md)


<a id="個人與群組-repository"></a>

```mermaid
flowchart LR
  h["個人與群組 Repository"]
  h --- n0["Repository 保存工程資料並管理庫內工區。跨庫<br/>引用與專案管理見 Project 功能規格。"]
  n0 --- n0l0["庫內工區"]
  click n0l0 "spaces.md"
  n0 --- n0l1["Project 功能規格"]
  click n0l1 "../project/project.md"
```

[Repository Insights 與 Pulse](insights.md) · [工程知識庫 Wiki](wiki.md)

<a id="範圍與資料"></a>

```mermaid
flowchart LR
  h["範圍與資料"]
  h --- n0["每個 LINE 群組一個 Bot 管理的空間，具有獨立 <br/>id 與 ownerId。 個人或群組由 LINE 事件綁定，<br/>目前未提供手動建庫、換綁或刪除入口。Bot 管理<br/>容器不授予使用者審核或讀取他人出勤的權限。"]
  h --- n1["已驗簽的 LINE follow 事件用 UserID 建立個人<br/>庫，並寫入本人候選清單。"]
  h --- n2["資料位於 channels/{channelId}/repositories/{<br/>id}，保存 id、ownerId、name、status、version<br/>、createdAt、updatedAt；內部另存最後事件時間<br/>及 ID。 repositoryOwnerBindings/{ownerId} 與<br/>容器在同一交易建立；ID 隨機產生，重送與併發<br/>沿用同一擁有者的綁定。 ACTIVE / INACTIVE 表<br/>示 Bot 群組入口狀態。離開保留資料，再加入沿<br/>用 ID。舊事件不得覆寫新狀態；同時間停用優先<br/>。版本只在公開狀態或名稱改變時遞增。 名稱是<br/>加入或首次互動時取得的群組名稱快照，查詢失敗<br/>時允許空白，後續事件會重試補齊；UI 顯示「群<br/>組空間」。不宣稱即時同步群組更名。"]
```

<a id="操作契約"></a>

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["欄位：Actor / Resource / Action<br/>觀察群組事件、建立／更新：驗證簽章的 Bot <br/>webhook；群組對應的 Repository；Observe"]
  h --- n1["欄位：Context / Authority<br/>觀察群組事件、建立／更新：Messaging API <br/>channel 範圍、事件時間與 ID；只接受 active <br/>模式群組加入、離開、訊息、postback、成員加入<br/>事件"]
  h --- n2["欄位：Policy / Decision<br/>觀察群組事件、建立／更新：groupId 格式有效；<br/>交易內讀取唯一綁定與現有容器，避免重複及舊事<br/>件覆寫；不消耗原事件，後續摘要與出勤照常執行"]
  h --- n3["欄位：Result<br/>觀察群組事件、建立／更新：原子建立／更新容器<br/>與使用者候選索引；不搬遷或改寫既有摘要、出勤<br/>資料"]
  h --- n4["欄位：重送／併發<br/>觀察群組事件、建立／更新：唯一綁定交易、固定<br/>新 ID 於重試外產生；相同事件不增加版本"]
  h --- n5["欄位：Event / Audit<br/>觀察群組事件、建立／更新：保留最後來源事件 <br/>ID、來源時間及公開更新時間；不保存 token、訊<br/>息內容或個人 profile"]
```

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["欄位：Actor / Resource / Action<br/>列表／詳情查詢：現有 Cookie Session 或 <br/>Bearer 驗證的使用者；Repository；List / Get"]
  h --- n1["欄位：Context / Authority<br/>列表／詳情查詢：候選來自已驗簽的事件。個人庫<br/>比對登入者與 ownerId；群組庫每次以 LINE <br/>group-member-profile 核驗當下成員資格"]
  h --- n2["欄位：Policy / Decision<br/>列表／詳情查詢：只顯示 ACTIVE 且核驗成功的容<br/>器；不存在、停用或無權詳情統一 404；LINE 非 <br/>200/404 或無效 profile 為 503，不能推定有權"]
  h --- n3["欄位：Result<br/>列表／詳情查詢：不可快取的基本資料；沒有摘要<br/>內容、他人出勤或管理能力"]
  h --- n4["欄位：重送／併發<br/>列表／詳情查詢：讀取無副作用；成員關係為查詢<br/>當下的外部觀察，不宣稱跨 LINE/Firestore 原子<br/>快照"]
  h --- n5["欄位：Event / Audit<br/>列表／詳情查詢：不為一般讀取另存個人資料或事<br/>件"]
```

<a id="api-與前端"></a>

```mermaid
flowchart LR
  h["API 與前端"]
  h --- n0["GET /api/v1/repositories?cursor=...：一次讀<br/>取本人最多十筆互動候選，再核驗目前資格；可能<br/>回傳空頁及 nextCursor，需繼續下一頁。"]
  h --- n1["GET /api/v1/repositories/{id}：核驗本人或當<br/>下群組成員資格；格式錯誤／不存在／無權均不揭<br/>露資料。"]
  h --- n2["API 契約：Repository，由 OpenAPI 入口 組合；<br/>生成 Go/TS DTO，不手改產生碼。"]
  n2 --- n2l0["Repository"]
  click n2l0 "../../../api/contract/repository/repositories.yaml"
  n2 --- n2l1["OpenAPI 入口"]
  click n2l1 "../../../api/contract/openapi.yaml"
  h --- n3["UI：側欄「空間」可展開已驗證使用者可開啟的空<br/>間，選取後在 <br/>/miniapp/repositories?selected=... 顯示既有<br/>詳情。原頁不再重複顯示另一份列表；既有 <br/>?id=... 及 repository／liff.state 群組入口仍<br/>相容。"]
  h --- n4["側欄一般選取使用乾淨 selected 參數，不沿用打<br/>卡 launch／其他表單 query；既有 id／<br/>repository 群組專用入口只讀指定詳情，不讀一<br/>般列表、不於失敗後退回其他群組。這些參數只在<br/>空間頁解析，不把其他功能的 id 當空間 ID。"]
  h --- n5["列表與詳情各自取消舊請求，並以 Session epoch<br/>、本人及作用域隔離可見狀態；登出／重新驗證立<br/>即隱藏舊資料，即使 transport 忽略取消也不得<br/>跨身分顯示。點選空間仍重新呼叫詳情 API 核驗<br/>，不以列表快取授權。"]
  h --- n6["手機選取空間後關閉抽屜；定位頁保留正式空間頁<br/>連結，不建立虛假的本人空間資料。無 <br/>JavaScript 時提供原路由連結及啟用提示。"]
  h --- n7["出勤連結只帶既有 group 查詢條件，不新增授權<br/>；摘要仍回原群組操作。"]
  h --- n8["從成員加入、本人群組訊息或 postback 建立列表<br/>候選。未互動的既有群組成員需先互動；這不是完<br/>整群組成員清單。"]
```

<a id="空間載入時序"></a>

## 空間載入時序

現行實作

```mermaid
sequenceDiagram
  participant UI as 側欄或空間頁
  participant Context as RepositoryContext
  participant API as Repository API
  alt 一般功能頁首次展開側欄或直接進入一般空間頁
    UI->>Context: 載入空間列表
    Context->>API: 查詢有權開啟的空間
  else 群組專用入口或選取空間
    UI->>Context: 載入指定詳情
    Context->>API: 重新核驗該空間
  end
  API-->>Context: 回應
  Context-->>UI: 僅套用目前本人、Session epoch 與作用域的回應
```

```mermaid
flowchart LR
  h["空間載入時序"]
  h --- n0["business/repository/RepositoryContext.tsx 共<br/>用列表、分頁與詳情載入。空頁仍有 nextCursor <br/>時保留「載入更多空間」。讀取失敗時清除舊項目<br/>並提供重試，候選或停用空間不得開啟。身分改變<br/>時的共用清理規則見頁面與身分生命週期。"]
  click n0 "../../architecture/runtime/react-lifecycle.md#頁面與身分生命週期"
```

<a id="正式空間-ui-基礎雛形"></a>

```mermaid
flowchart LR
  h["正式空間 UI 基礎雛形"]
  h --- n0["正式空間沿用上述 List / Get 操作契約與身分隔<br/>離，以寬版卡片呈現已授權空間的基本資料與本人<br/>出勤入口。 正式頁內導覽由 <br/>frontend/miniapp/src/business/repository/navigation.ts<br/> 定義；Repository內容維持原有問題、討論、標<br/>籤及里程碑。專案改走獨立側欄／<br/>/miniapp/projects，舊入口僅導向帶明確Space／<br/>來源識別的新頁，不再在Repository View掛載<br/>Project業務。 總覽使用真實空間資料，問題已接<br/>入 問題第一階段，討論已接入 討論與留言，標籤<br/>目錄已接入 標籤規格；里程碑、專案及洞察<br/>依里程碑、Project 與 Insights 規格接入；公開稽核仍明<br/>示尚未開放，不回填示範資料。載入、核驗失敗、<br/>登出時不呈現空間內容。 既有 API、Bot 管理、<br/>成員核驗及共用編號規則不變。問題與討論／留言<br/>依各自獨立規格接入；標籤已接入問題／討論作者<br/>的分類操作；完成界線依 基礎雛形，里程碑、專<br/>案及洞察依本次授權另完成最小切片，不自動展開<br/>其他管理功能。 總覽為讀取畫面基礎整合；問題<br/>驗證另依功能規格，不代表其他協作業務或真實 <br/>LINE 裝置已驗收。"]
  n0 --- n0l0["問題第一階段"]
  click n0l0 "../issue/design.md"
  n0 --- n0l1["討論與留言"]
  click n0l1 "../discussion/design.md"
  n0 --- n0l2["標籤規格"]
  click n0l2 "labels.md"
  n0 --- n0l3["里程碑"]
  click n0l3 "milestones.md"
  n0 --- n0l4["Project"]
  click n0l4 "../project/project.md"
  n0 --- n0l5["洞察（Insights）"]
  click n0l5 "insights.md"
  n0 --- n0l6["基礎雛形"]
  click n0l6 "acceptance.md"
```

<a id="啟用與遷移限制"></a>

```mermaid
flowchart LR
  h["啟用與遷移限制"]
  h --- n0["沿用 GOOGLE_CLOUD_PROJECT、<br/>FIRESTORE_DATABASE_ID、Messaging API 設定；<br/>缺少設定時查詢回 503。 MINI App 與 Messaging<br/> API 使用者識別須在相同 LINE provider 下對應<br/>；無法匹配時不能通過成員核驗。 既有群組收到<br/>下一筆支援的可信事件時建立，不執行正式資料掃<br/>描或批次遷移。Summary / Attendance 歷史資料<br/>仍以 groupId 保存；容器原始切片只建立容器、<br/>解析介面與讀取 UI；問題、討論及標籤各依功能<br/>規格寫入，不搬遷既有出勤或摘要資料。"]
```

<a id="查證與驗收"></a>

```mermaid
flowchart LR
  h["查證與驗收"]
  h --- n0["LINE 官方來源 main：<br/>dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54，查<br/>閱 group chats、Messaging API reference 的 <br/>join/leave/group-member-profile，確認 404 不<br/>授權及成員 profile 的 userId。"]
  h --- n1["Context7 /googleapis/google-cloud-go：<br/>Firestore RunTransaction 重試及先讀後寫；對<br/>照專案 firestore v1.29.2。Google Developer <br/>Knowledge 未在此聊天提供工具，使用官方 Go <br/>client 文件；未做雲端部署或資料操作。"]
  h --- n2["Context7 /vercel/next.js/v16.4.5 與安裝版 <br/>static-export 指引：沿用 Pages Router、客戶<br/>端載入及靜態輸出。"]
  h --- n3["Graphify 已定位 webhook、server、identity 及<br/>既有群組查詢；Repomix 使用本次限定範圍產物 <br/>/tmp/repository-before.xml，未涵蓋的指引與文<br/>件直接閱讀。"]
  h --- n4["本機檢查結果於交付訊息列出；LINE 真實群組事<br/>件、正式 Firestore 併發與 iOS/Android 登入需<br/>另行環境驗收，不能以建置通過替代。"]
  h --- n5["本次本機驗證：前端靜態建置與型別檢查、12 個<br/>既有前端測試、OpenAPI 產生碼同步／相容性、設<br/>定檢查、Go vet／既有 race 測試／build／lint<br/>／格式與漏洞檢查通過。獨立暫存 smoke 使用受<br/>控 store／身分與 LINE transport，確認未登入<br/>、非成員、停用、錯誤 cursor、profile 不符與 <br/>404／429 的拒絕語義；不代表 Firestore 真實交<br/>易或 LINE 整合驗收。未新增正式測試案例或工具<br/>。"]
  h --- n6["LINE「選單 → 請選擇功能」與功能說明提供「空<br/>間」URI 入口，以既有 LIFF URL 加 <br/>repositories=1，首頁解析一般 query 或 <br/>liff.state 後導向空間頁。"]
  h --- n7["群組空間入口由可信 groupId 查得 Repository <br/>ID，連結帶 repositories=1&amp;repository=&lt;id&gt;。<br/>一般 query 與 liff.state 均保留目標 ID；專用<br/>入口只查詳情，重新載入仍查同一 ID，失敗不退<br/>回列表。私訊入口維持本人列表。個人庫比對本人<br/>，群組庫以 ownerId 即時核驗成員；連結不是授<br/>權憑證，修改 ID 不授予其他群組權限。"]
  h --- n8["群組專用入口本機驗證：以實際 Go 靜態頁路由、<br/>受控 LINE SDK／API 回應檢查一般 query、<br/>liff.state 與 404 情境；三者均只請求指定 <br/>Repository 詳情，重新載入不切換到列表，390px<br/> 無水平溢出。這不取代 LINE 裝置驗收。"]
  h --- n9["工程項目新增與追蹤依 Objects 規格，與「問題<br/>」業務分開；不自動匯入 PDF 或授予他人編輯權<br/>。"]
  click n9 "../object/design.md"
```

<a id="後續模型設計"></a>

```mermaid
flowchart LR
  h["後續模型設計"]
  h --- n0["內部共用編號基礎見 Repository 共用資源編號。<br/>問題與討論建立已接入，留言不占編號；本頁 Bot<br/> 管理、公開查詢及成員核驗規則保持原樣。"]
  click n0 "numbering.md"
  h --- n1["Repository 的實體、查詢、集合與 Connection <br/>分頁命名，以及 Issue／Discussion 的歸屬規劃<br/>，見 業務模型設計。該頁為後續開發參考，不變<br/>更本頁既有 REST、成員核驗與 Bot 管理模式。"]
  click n1 "../../proposals/platform/domain-model.md"
```

<a id="decision-flow"></a>

## 寫入庫內受限資源的判定流程

```mermaid
flowchart TD
    start["寫入庫內受限資源"] --> gate{"目前擁有者與啟用狀態符合？"}
    gate -->|是| ok["允許繼續資源操作"]
    gate -->|否| stop["拒絕沿用舊授權"]
```

[入口適配器](../../../backend/internal/business/repository/adapter/http/routes.go) · [判定依據](../../../backend/internal/business/repository/domain/entity/repository.go) · [應用層輸出介面](../../../backend/internal/business/repository/application/port/repository.go) · [相關測試](../../../backend/internal/business/repository/adapter/firestore/labels_test.go)

[適用工作指引](../AGENTS.md) · [上層導覽](../README.md)
