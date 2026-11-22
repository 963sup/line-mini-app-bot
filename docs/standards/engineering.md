# 開發原則與決策

<a id="開發原則與決策"></a>

```mermaid
flowchart LR
  h["開發原則與決策"]
  h --- n0["預設：模組化單體＋輕量 Hexagonal＋Go 慣例。 <br/>&gt; 先依本頁選做法，再到整合文件查技術細節。"]
```

<a id="1-原則如何落地"></a>

```mermaid
flowchart LR
  h["1. 原則如何落地"]
  h --- n0["方法：Go Idioms<br/>採用程度：必要<br/>實際做法：小介面、組合、明確錯誤處理、傳遞 <br/>context 並限制外部呼叫時間"]
  h --- n1["方法：KISS／YAGNI<br/>採用程度：必要<br/>實際做法：先維持一個服務；有當前需求與證據後<br/>再增加功能或抽象"]
  h --- n2["方法：模組化單體<br/>採用程度：預設架構<br/>實際做法：按業務能力拆模組；單一容器部署至 <br/>Cloud Run"]
  h --- n3["方法：輕量 Hexagonal<br/>採用程度：外部依賴邊界採用<br/>實際做法：隔離 HTTP、LINE SDK 與 Firestore，<br/>業務不依賴 SDK 型別"]
  h --- n4["方法：DDD<br/>採用程度：選擇性採用<br/>實際做法：先統一業務用語與邊界；規則複雜後才<br/>引入完整領域模型"]
  h --- n5["方法：明確依賴注入<br/>採用程度：預設做法<br/>實際做法：建構函式接收依賴，在 main 組裝，通<br/>常不需要 DI 框架"]
  h --- n6["方法：DRY<br/>採用程度：依知識判斷<br/>實際做法：共用同一業務規則；外觀相似但變更原<br/>因不同的程式可分開保留"]
  h --- n7["方法：可觀測性與測試<br/>採用程度：依變更風險<br/>實際做法：保留 slog 與必要操作說明，執行既有<br/>檢查並補足關鍵回歸證據"]
```

<a id="2-遇到問題直接選"></a>

```mermaid
flowchart LR
  h["2. 遇到問題直接選"]
  h --- n0["當你要做…：Webhook 與 MINI App 提供相同操作<br/>優先做法：兩個入口呼叫同一使用案例"]
  h --- n1["當你要做…：新增 Firestore 讀寫或 LINE 發送<br/>優先做法：在業務使用端定義所需的小介面，由 <br/>adapter 實作"]
  h --- n2["當你要做…：新增只有一處使用的簡單判斷<br/>優先做法：先寫清楚的函式，等責任與需求出現再<br/>抽象"]
  h --- n3["當你要做…：新增業務規則<br/>優先做法：在業務層實作，先驗證成功與拒絕案例"]
  h --- n4["當你要做…：同時更新多筆相關資料<br/>優先做法：評估 Firestore transaction，外部副<br/>作用放在 callback 外"]
  h --- n5["當你要做…：保證資料更新後最終送出通知<br/>優先做法：定義失敗語義；有可靠投遞需求時再設<br/>計持久化工作／outbox"]
  h --- n6["當你要做…：改善效能<br/>優先做法：先量測，再用 benchmark／pprof 驗證"]
  h --- n7["當你要做…：引入新架構或框架<br/>優先做法：說明目前問題、簡單方案的限制與驗證<br/>方式，再記錄 ADR"]
```

<a id="3-每次功能交付順序"></a>

## 3. 每次功能交付順序

開發與維運指引

```mermaid
flowchart LR
  A["定義入口、可信身分、輸入與預期結果"]
  A --> B["在使用案例實作業務規則與授權"]
  B --> C["接上 adapter 並處理逾時與失敗"]
  C --> D["驗證成功、拒絕、重複請求與下游失敗"]
  D --> E["核對日誌可追查且排除 token、機密與不必要個資"]
```

```mermaid
flowchart LR
  h["3. 每次功能交付順序"]
  h --- n0["業務變更的查證與交付證據依業務操作符合性流程<br/>。"]
  click n0 "../development/workflow.md#business-operation-compliance"
```

<a id="4-邊界檢查"></a>

```mermaid
flowchart LR
  h["4. 邊界檢查"]
  h --- n0["Webhook：先驗證原始 body 的簽章，處理重送冪<br/>等性，詳見 LINE。"]
  click n0 "../integrations/line/capabilities.md"
  h --- n1["MINI App：後端驗證 LINE 身分，再檢查資源存取<br/>權；不信任前端 user ID。"]
  h --- n2["Firestore：server SDK 的授權由後端與 IAM 負<br/>責；交易 callback 可能重跑，詳見 Firestore。"]
  click n2 "../integrations/firestore/connection.md"
  h --- n3["並行工作：明確管理取消與生命週期；需要可靠完<br/>成的工作使用持久化機制。"]
  h --- n4["目前先不引入微服務、完整 CQRS、Event <br/>Sourcing 或通用 Repository 框架。 當獨立部署<br/>、讀寫差異或事件稽核需求成立時，再重新評估。"]
  h --- n5["實作細節見 Go 慣例；驗證方式見 驗收矩陣。"]
  n5 --- n5l0["Go 慣例"]
  click n5l0 "go.md"
  n5 --- n5l1["驗收矩陣"]
  click n5l1 "../development/acceptance.md#變更驗收矩陣"
  h --- n6["目標套件配置見 目錄規劃，依賴方向見 依賴規則<br/>。"]
  n6 --- n6l0["目錄規劃"]
  click n6l0 "../architecture/runtime/code-layout.md"
  n6 --- n6l1["依賴規則"]
  click n6l1 "../architecture/runtime/dependencies.md"
```

<a id="資源導向業務操作"></a>

```mermaid
flowchart LR
  h["資源導向業務操作"]
  h --- n0["新增或修改功能須依 Business Operation 模型與<br/>操作契約 定義身分、資源、行為、授權、領域規<br/>則與結果。 完整定義集中於該頁；本頁既有模組<br/>邊界、簡單優先與漸進演進原則繼續適用。"]
  click n0 "../contracts/business-operations.md"
```
