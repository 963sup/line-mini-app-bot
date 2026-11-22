# 共用系統基礎

修改跨模組行為時，依下列問題找契約與程式入口。單一功能的規則由[領域規格](../domains/README.md)維護。需要劃分模型、交易或 Port 時，從 [DDD 問題索引](ddd/README.md#按問題查閱)選讀。

## 現行架構與決策

### 現行結構與規則

| 問題 | 來源 |
| --- | --- |
| 哪個入口處理請求？改動會牽涉哪些模組？ | [服務入口](runtime/overview.md) · [從需求定位修改範圍](runtime/code-layout.md#從需求定位修改範圍) |
| 核心、適配器與組裝在哪裡？ | [後端功能邊界](runtime/backend-boundaries.md) · [依賴與契約歸屬](runtime/dependencies.md) |
| 這項操作由誰授權？哪些規則必須一起成立？ | [業務操作模型](contracts/business-operations.md) · [工程原則](../standards/engineering.md) |

### 任務準備與按需參考

從[任務索引](../README.md#按任務開始)選取範圍，依[驗收矩陣](../development/acceptance.md#變更驗收矩陣)決定必要證據。早期[平台交付階段](../proposals/platform/delivery.md)記錄當時的切片與待決策項目，功能現況查所屬契約。

### 接手時的證據與用語

追蹤需求時，將需求 ID 對到用例、實際檔案或函數與代表測試。成功與拒絕案例的 JSON 放在測試的 testdata。actor 是經驗證操作者，target 是發送目標；accepted 表示 API 接受，completed 表示本地處理完成，兩者均未證明訊息已讀。核對每類錯誤的 HTTP、重試與觀測語意；未知錯誤不預設可以重試。

工具版本查 [mise](../../mise.toml)，Go SDK 依賴查 [go.mod](../../backend/go.mod)。畫圖時區分執行、依賴及資料方向；Graphify 描述程式結構，Repomix 提供來源上下文，驗收另需行為證據。

### 歷史查核來源（2026-11-07）

當時以 Context7 查核 Go 命名與建構函數、LINE event source／mode、retry key、Reply，以及 Firestore 交易重試與刪除。來源 ID 為 `/websites/go_dev_doc`、`/line/line-openapi`、`/websites/cloud_google_firestore_native`。遇到 token 期限、資格或客戶端差異，依目前 endpoint 文件及實測確認。

<a id="architecture-decision-records"></a>

### 已接受的架構決策

[ADR 0002](../decisions/0002-baseline.md) 定義應用基線，部署部分由 [ADR 0003](../decisions/0003-cloud-run-primary.md) 取代；[ADR 0004](../decisions/0004-monorepo.md) 說明 Monorepo 邊界。新增重大決策時，在[決策目錄](../decisions/README.md)記下背景、選項、狀態、影響與驗證，標記已被取代的決策並保留歷史。

## 相鄰契約

### HTTP 契約

查[公開 API 與身分契約](contracts/http.md)及[跨領域契約索引](contracts/README.md)。實際變更 API 的命令在 [API 工作流程](../../api/README.md)。

### 身分與登入

[身分索引](identity/README.md) · [Go browser Session、Firestore 與續期](identity/browser-session.md)

### 共用前端

[React 頁面生命週期](runtime/react-lifecycle.md) · [執行架構索引](runtime/README.md)

### 資料與儲存

先讀[功能規格](../domains/README.md)，再定位該領域的 Firestore adapter。[儲存索引](storage/README.md)導向[資料配置](storage/data-model.md)、[交易](storage/consistency.md)、[查詢與索引](storage/querying.md)、[生命週期](storage/lifecycle.md)及[Firestore 整合](storage/firestore.md)；[資料契約提案](../proposals/firestore-data-contracts.md)描述規劃中的 schemaVersion、Inbox／Outbox 與保存策略。

### 外部整合

<a id="文件索引"></a>

[整合索引](integrations/README.md) · [LINE 接入邊界](integrations/line.md) · [LINE 環境變數](integrations/line-environment.md) · [呈現方式](integrations/line-presentation-options.md) · [瀏覽器定位](integrations/web-geolocation.md)

### 部署與營運

從 [Cloud Run 建置入口](../guides/release/cloud-run.md#建置入口)、[發布與回復](../guides/release/release-acceptance.md#切換觀察回復)或[故障診斷](../guides/release/troubleshooting.md)進入。需要分別查儲存庫設定、雲端實際設定、建置結果與服務驗收。

### 可靠性設計參考

[Webhook、Inbox 與 Outbox 提案](../proposals/webhook-delivery/design.md)用於討論重試及持久工作；是否已實作以[現行 HTTP 契約](contracts/http.md)及對應功能為準。

## 歸屬與維護

修改架構、外部整合或營運文件時，依[本層指引](AGENTS.md)與[索引維護規則](../AGENTS.md#主題子目錄)更新受影響的入口。常用圖見[專案閱讀入口](../README.md#日常開發看哪張圖)。
