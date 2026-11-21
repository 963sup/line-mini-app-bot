# 共用系統基礎

[跨領域契約](contracts/README.md) · [執行架構](runtime/README.md) · [身分與工作階段](identity/README.md) · [外部服務整合](integrations/README.md) · [資料儲存](storage/README.md)

## 用途與範圍

保存系統架構、契約、整合與執行邊界的說明。

## 架構與責任邊界

跨模組設計放在此處，單一領域的規則由領域文件維護。

## 閱讀與修改入口

[DDD 戰略與戰術設計、上下文映射及六邊形架構](ddd/README.md)

先查閱、[本目錄工作指引](AGENTS.md)。

## 相關規格

[文件總入口](../README.md) · [HTTP 與身分契約](contracts/http.md)


<a id="共用系統基礎"></a>

```mermaid
flowchart LR
  h["共用系統基礎"]
  h --- n0["只讀任務涉及的共用能力。業務規則在 features<br/>，未來平台設計在 planning。"]
  n0 --- n0l0["features"]
  click n0l0 "../domains/README.md"
  n0 --- n0l1["planning"]
  click n0l1 "../proposals/README.md"
```

```mermaid
flowchart LR
  h["共用系統基礎"]
  h --- n0["分類：architecture<br/>內容：現行架構、模組責任、依賴與已接受決策<br/>何時閱讀：定位程式、調整邊界時"]
  click n0 "#現行架構與決策"
  h --- n1["分類：identity<br/>內容：Session 與可信身分<br/>何時閱讀：修改登入、續期與認證時"]
  click n1 "#身分與登入"
  h --- n2["分類：api<br/>內容：共用 HTTP 與身分契約<br/>何時閱讀：修改入口、安全與錯誤語意時"]
  click n2 "#http-契約"
  h --- n3["分類：frontend<br/>內容：React 共用生命週期<br/>何時閱讀：修改前端非同步請求與頁面狀態時"]
  click n3 "#共用前端"
  h --- n4["分類：persistence<br/>內容：Firestore 整合與資料設計參考<br/>何時閱讀：修改儲存與交易時；先確認各頁狀態"]
  click n4 "#資料與儲存"
  h --- n5["分類：integrations<br/>內容：LINE 與瀏覽器能力整合<br/>何時閱讀：修改平台接入或定位時"]
  click n5 "#外部整合"
  h --- n6["分類：reliability<br/>內容：事件、狀態機與可靠發送設計<br/>何時閱讀：涉及重試或持久工作時；不預設所有設<br/>計已實作"]
  click n6 "#可靠性設計參考"
  h --- n7["分類：operations<br/>內容：Cloud Run、發布驗收與排錯<br/>何時閱讀：涉及部署與真實環境問題時"]
  click n7 "#部署與營運"
```

## 共用系統基礎

[文件總入口](../README.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="http-契約"></a>

## HTTP 契約

[HTTP 與身分契約](contracts/http.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="現行架構與決策"></a>

```mermaid
flowchart LR
  h["現行架構與決策"]
  h --- n0["先以需求導航找規格與實作；架構文件按問題查閱<br/>，不作為每個任務必讀清單。"]
  click n0 "runtime/code-layout.md#從需求定位修改範圍"
```

<a id="現行結構與規則"></a>

```mermaid
flowchart LR
  h["現行結構與規則"]
  h --- n0["文件：架構總覽<br/>回答的問題：目前服務如何執行、有哪些入口？"]
  click n0 "runtime/overview.md"
  h --- n1["文件：需求導航與目錄責任<br/>回答的問題：需求應修改哪個模組？"]
  click n1 "runtime/code-layout.md"
  h --- n2["文件：後端功能邊界<br/>回答的問題：核心、adapter 與組裝如何分工？"]
  click n2 "runtime/backend-boundaries.md"
  h --- n3["文件：依賴與契約歸屬<br/>回答的問題：技術版本、共用規則與 import 邊界<br/>在哪裡？"]
  click n3 "runtime/dependencies.md"
  h --- n4["文件：開發原則<br/>回答的問題：如何選擇簡單且符合現行邊界的做法<br/>？"]
  click n4 "../standards/engineering.md"
  h --- n5["文件：業務操作模型<br/>回答的問題：如何描述授權、規則與操作結果？"]
  click n5 "contracts/business-operations.md"
  h --- n6["文件：ADR<br/>回答的問題：哪些重大決策已接受或被取代？"]
  click n6 "#architecture-decision-records"
```

<a id="任務準備與按需參考"></a>

```mermaid
flowchart LR
  h["任務準備與按需參考"]
  h --- n0["依任務選讀；完成條件見驗收矩陣，原始交付順序<br/>見架構基線。缺產品決策時先查當前功能規格。"]
  n0 --- n0l0["依任務選讀"]
  click n0l0 "../README.md#按任務開始"
  n0 --- n0l1["驗收矩陣"]
  click n0l1 "../development/acceptance.md#變更驗收矩陣"
  n0 --- n0l2["架構基線"]
  click n0l2 "../decisions/0002-baseline.md"
  h --- n1["實作契約與待決策：保留早期切片與決策閘門；個<br/>別功能以最新功能契約為準，不將舊待決策表當作<br/>目前完成狀態。"]
  click n1 "../proposals/platform/delivery.md"
  h --- n2["未來規劃與提案：平台擴充、目標業務模型與舊交<br/>付階段已移出現行架構目錄。"]
  click n2 "../proposals/README.md"
```

<a id="接手時的證據與用語"></a>

```mermaid
flowchart LR
  h["接手時的證據與用語"]
  h --- n0["具體功能 happy path／拒絕範例與 JSON <br/>fixtures，放測試 testdata，避免在多頁重複維<br/>護。"]
  h --- n1["一張需求 ID → Use Case → 檔案／函數 → 測試的<br/>追溯表；功能確定後填真實路徑，不預填不存在證<br/>據。"]
  h --- n2["用詞表：actor=經驗證操作者；target=發送目標<br/>；accepted=API 接受；completed=本地處理完成<br/>，均不等於已讀。"]
  h --- n3["各錯誤的重試／HTTP／觀測映射；未知不得默認可<br/>重試。"]
  h --- n4["版本與來源：工具在 mise、SDK 在 go.mod；文件<br/>記查核來源，不另設版本真相。"]
  h --- n5["圖要標 execution/import/data 關係；一圖一問<br/>題；規劃圖不可混進 Graphify 現況圖冒充已實作<br/>。"]
```

<a id="歷史查核來源2026-11-07"></a>

```mermaid
flowchart LR
  h["歷史查核來源（2026-11-07）"]
  h --- n0["Context7：/websites/go_dev_doc（MixedCaps／<br/>建構函數）、/line/line-openapi（事件 source<br/>／mode、retry key、Reply）、<br/>/websites/cloud_google_firestore_native（交<br/>易重試、資料刪除），2026-11-07。 Context7 回<br/>傳片段不是完整官方 reference；未回傳的 token<br/> 期限、資格與 client-specific 行為必須再核對<br/> endpoint 文件與實測。 Graphify AST 只描述可<br/>擷取程式關係，Markdown 規劃用 Mermaid 與連結<br/>；Repomix 用來提供可裁剪上下文，不證明應用驗<br/>收。"]
```

<a id="相鄰契約"></a>

```mermaid
flowchart LR
  h["相鄰契約"]
  h --- n0["功能規格 · HTTP · Session · 資料 · 前端 · 驗<br/>收矩陣"]
  n0 --- n0l0["功能規格"]
  click n0l0 "../domains/README.md"
  n0 --- n0l1["HTTP"]
  click n0l1 "#http-契約"
  n0 --- n0l2["Session"]
  click n0l2 "#身分與登入"
  n0 --- n0l3["資料"]
  click n0l3 "#資料與儲存"
  n0 --- n0l4["前端"]
  click n0l4 "#共用前端"
  n0 --- n0l5["驗收矩陣"]
  click n0l5 "../development/acceptance.md"
  h --- n1["本目錄指引 · 父層指引 · 文件總入口"]
  n1 --- n1l0["本目錄指引"]
  click n1l0 "AGENTS.md#architecture"
  n1 --- n1l1["父層指引"]
  click n1l1 "../AGENTS.md"
  n1 --- n1l2["文件總入口"]
  click n1l2 "../README.md"
  h --- n2["現行實作圖索引 · 未來設計圖。"]
  n2 --- n2l0["現行實作圖索引"]
  click n2l0 "../diagrams.md#現行實作"
  n2 --- n2l1["未來設計圖"]
  click n2l1 "../diagrams.md#未來設計"
```

<a id="architecture-decision-records"></a>

```mermaid
flowchart LR
  h["Architecture decision records"]
  h --- n0["Accepted, deployment portion superseded by <br/>ADR 0003: application baseline"]
  click n0 "../decisions/0002-baseline.md"
  h --- n1["Accepted: Cloud Run deployment"]
  click n1 "../decisions/0003-cloud-run-primary.md"
  h --- n2["Add a numbered record when a consequential <br/>technical choice is made. Include status, <br/>problem/context, options, decision, <br/>consequences and validation. Reference <br/>affected modules and official sources. Mark <br/>replaced decisions as superseded rather than<br/> silently rewriting history. Do not invent <br/>completed implementation or deployment <br/>evidence."]
  h --- n3["ADR 0004：前後端與 API 契約分離的 Monorepo"]
  click n3 "../decisions/0004-monorepo.md"
  h --- n4["現行實作圖索引 · 未來設計圖。"]
  n4 --- n4l0["現行實作圖索引"]
  click n4l0 "../diagrams.md#現行實作"
  n4 --- n4l1["未來設計圖"]
  click n4l1 "../diagrams.md#未來設計"
```

<a id="共用前端"></a>

## 共用前端

[React 前端生命週期](runtime/react-lifecycle.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="身分與登入"></a>

## 身分與登入

[Go browser Session 與 Firestore](identity/browser-session.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="外部整合"></a>

```mermaid
flowchart LR
  h["外部整合"]
  h --- n0["LINE、Firestore 與瀏覽器定位等整合契約及限制<br/>。"]
```

<a id="文件索引"></a>

## 文件索引

[Firestore integration](storage/firestore.md)

[LINE 環境變數規格](integrations/line-environment.md)

[LINE 呈現方式選型](integrations/line-presentation-options.md)

[LINE 整合：能力與邊界](integrations/line.md)

[Web Geolocation API 整合契約](integrations/web-geolocation.md)

<a id="歸屬與維護"></a>

```mermaid
flowchart LR
  h["歸屬與維護"]
  h --- n0["外部服務契約與專案功能規則分開；遵循根層查證<br/>政策，不在文件存放 token、憑證或機密。 新增<br/>或搬移文件時，依文件維護指引更新索引。"]
  click n0 "../AGENTS.md#主題子目錄"
  h --- n1["本目錄指引 · 父層指引 · 文件總索引"]
  n1 --- n1l0["本目錄指引"]
  click n1l0 "AGENTS.md#integrations"
  n1 --- n1l1["父層指引"]
  click n1l1 "../AGENTS.md"
  n1 --- n1l2["文件總索引"]
  click n1l2 "../README.md"
  h --- n2["現行實作圖索引 · 未來設計圖。"]
  n2 --- n2l0["現行實作圖索引"]
  click n2l0 "../diagrams.md#現行實作"
  n2 --- n2l1["未來設計圖"]
  click n2l1 "../diagrams.md#未來設計"
```

<a id="部署與營運"></a>

## 部署與營運

Cloud Run 部署、發布驗收與故障排查。

<a id="文件索引"></a>


Cloud Run 部署 · 建置入口圖

[Cloud Run 部署](../guides/release/cloud-run.md)

[建置入口圖](../guides/release/cloud-run.md#建置入口)

Monorepo 整合、發布與回復 · 切流與回復圖

[Monorepo 整合、發布與回復](../guides/release/release-acceptance.md)

[切流與回復圖](../guides/release/release-acceptance.md#切換觀察回復)

[Troubleshooting](../guides/release/troubleshooting.md)

<a id="歸屬與維護"></a>

```mermaid
flowchart LR
  h["歸屬與維護"]
  h --- n0["明確區分儲存庫設定、雲端實際設定、建置成功與<br/>服務驗收；文件修改不授權部署或外部寫入。 新<br/>增或搬移文件時，依文件維護指引更新索引。"]
  click n0 "../AGENTS.md#主題子目錄"
  h --- n1["本目錄指引 · 父層指引 · 文件總索引"]
  n1 --- n1l0["本目錄指引"]
  click n1l0 "AGENTS.md#operations"
  n1 --- n1l1["父層指引"]
  click n1l1 "../AGENTS.md"
  n1 --- n1l2["文件總索引"]
  click n1l2 "../README.md"
  h --- n2["開發與維運指引圖索引 · 未來設計圖。"]
  n2 --- n2l0["開發與維運指引圖索引"]
  click n2l0 "../diagrams.md#開發與維運指引"
  n2 --- n2l1["未來設計圖"]
  click n2l1 "../diagrams.md#未來設計"
```

<a id="資料與儲存"></a>

```mermaid
flowchart LR
  h["資料與儲存"]
  h --- n0["修改現行資料時先讀所屬功能規格，再定位該功能<br/>的 Firestore adapter；本目錄不提供涵蓋全部現<br/>行 collection 的資料字典。"]
  click n0 "../domains/README.md"
  h --- n1["Firestore 整合：連線、交易與整合注意事項。"]
  click n1 "storage/firestore.md"
  h --- n2["資料契約提案：尚未實作的 schemaVersion、<br/>Inbox／Outbox、保存與刪除設計；不作為現行資<br/>料庫全貌。"]
  click n2 "../proposals/firestore-data-contracts.md"
  h --- n3["共用系統入口 · 需求導航 · 未來規劃"]
  n3 --- n3l0["共用系統入口"]
  click n3l0 "README.md"
  n3 --- n3l1["需求導航"]
  click n3l1 "runtime/code-layout.md"
  n3 --- n3l2["未來規劃"]
  click n3l2 "../proposals/README.md"
  h --- n4["現行實作圖索引 · 未來設計圖。"]
  n4 --- n4l0["現行實作圖索引"]
  click n4l0 "../diagrams.md#現行實作"
  n4 --- n4l1["未來設計圖"]
  click n4l1 "../diagrams.md#未來設計"
```

<a id="可靠性設計參考"></a>

```mermaid
flowchart LR
  h["可靠性設計參考"]
  h --- n0["本目錄的設計文件屬於通用 Inbox／Outbox 與後<br/>續平台提案，不表示目前已有持久接收、租約 <br/>Worker 或完整可靠投遞能力。現行行為以功能規<br/>格及 HTTP 契約為準。"]
  n0 --- n0l0["功能規格"]
  click n0l0 "../domains/README.md"
  n0 --- n0l1["HTTP 契約"]
  click n0l1 "contracts/http.md"
  h --- n1["事件與發送狀態機提案"]
  click n1 "../proposals/webhook-delivery/design.md"
  h --- n2["只在任務涉及持久工作、重試或平台設計時載入；<br/>目前必要檢查見驗收矩陣。規劃入口"]
  n2 --- n2l0["驗收矩陣"]
  click n2l0 "../development/acceptance.md"
  n2 --- n2l1["規劃入口"]
  click n2l1 "../proposals/README.md"
  h --- n3["未來設計圖索引 · 對照現行實作。"]
  n3 --- n3l0["未來設計圖索引"]
  click n3l0 "../diagrams.md#未來設計"
  n3 --- n3l1["對照現行實作"]
  click n3l1 "../diagrams.md#現行實作"
```
