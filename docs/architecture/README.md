# 共用系統基礎

修改跨模組行為時，依下列問題找契約與程式入口。單一功能的規則由[領域規格](../domains/README.md)維護。需要劃分模型、交易或 Port 時，從 [DDD 問題索引](../references/ddd/README.md#按問題查閱)選讀。

## 現行架構與決策

### 現行結構與規則

| 問題 | 來源 |
| --- | --- |
| 哪個入口處理請求？改動會牽涉哪些模組？ | [服務入口](runtime/overview.md) · [從需求定位修改範圍](runtime/code-layout.md#從需求定位修改範圍) |
| 核心、適配器與組裝在哪裡？ | [後端功能邊界](runtime/backend-boundaries.md) · [依賴與契約歸屬](runtime/dependencies.md) |
| 這項操作由誰授權？哪些規則必須一起成立？ | [業務操作模型](../contracts/business-operations.md) · [工程原則](../standards/engineering.md) |

### 任務準備與按需參考

從[任務索引](../README.md#按任務開始)選取範圍，依[驗收矩陣](../development/tooling/testing/commands.md#變更驗收矩陣)決定必要證據。早期[平台交付階段](../proposals/platform/delivery.md)記錄當時的切片與待決策項目，功能現況查所屬契約。

<a id="architecture-decision-records"></a>

### 已接受的架構決策

[ADR 0002](../decisions/0002-baseline.md) 定義應用基線，部署部分由 [ADR 0003](../decisions/0003-cloud-run-primary.md) 取代；[ADR 0004](../decisions/0004-monorepo.md) 說明 Monorepo 邊界。新增重大決策時，在[決策目錄](../decisions/README.md)記下背景、選項、狀態、影響與驗證，標記已被取代的決策並保留歷史。

## 相鄰契約

### HTTP 契約

查[公開 API 與身分契約](../contracts/http.md)及[跨領域契約索引](../contracts/README.md)。實際變更 API 的命令在 [API 工作流程](../../api/README.md)。

### 身分與登入

[身分與 Session 契約](../contracts/identity.md)

### 共用前端

[React 頁面生命週期](runtime/react-lifecycle.md) · [執行架構索引](runtime/README.md)

### 資料與儲存

[資料儲存責任](data.md)說明領域與 adapter 的分工。查閱 [Firestore](../integrations/firestore/README.md) 或 [Cloud Storage](../integrations/storage/README.md) 的具體契約。

### 外部整合

<a id="文件索引"></a>

[外部整合索引](../integrations/README.md) · [LINE 整合](../integrations/line/README.md) · [瀏覽器定位](../integrations/geolocation/README.md)

### Google 服務選用

[Google Cloud 與 Maps 簡要規格](../integrations/google-cloud/README.md)依服務分頁，包含背景工作、functions、機密與 IAM、觀測、建置、AI、報表、Maps 及 reCAPTCHA。檔案儲存由 [Cloud Storage](../integrations/storage/cloud-storage.md)維護，既有 Firestore 與 Cloud Run 沿用原來源。各頁標示現況或導入條件，不代表候選服務已啟用。

### 部署與營運

從 [Cloud Run 建置入口](../operations/cloud-run.md#建置入口)、[發布與回復](../operations/release-acceptance.md#切換觀察回復)或[故障診斷](../operations/troubleshooting.md)進入。需要分別查儲存庫設定、雲端實際設定、建置結果與服務驗收。

### 可靠性設計參考

[Webhook、Inbox 與 Outbox 提案](../proposals/webhook-delivery/design.md)用於討論重試及持久工作；是否已實作以[現行 HTTP 契約](../contracts/http.md)及對應功能為準。

## 歸屬與維護

修改架構、外部整合或營運文件時，依[本層指引](AGENTS.md)與[索引維護規則](../AGENTS.md#主題子目錄)更新受影響的入口。常用圖見[專案閱讀入口](../README.md#日常開發看哪張圖)。
