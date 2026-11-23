# 事件觸發與流程協調

<a id="pubsub"></a>

## Pub/Sub

**狀態：多個消費者需要同一事件時導入。**

**用途與最小流程**

將已發生的業務事實分送通知、統計或搜尋等獨立消費者。

權威操作提交 → 可靠發布事件 → 各訂閱者驗證與去重 → 完成保存後確認訊息。

**資料、授權與失敗邊界**

事件帶 ID、種類、版本、來源範圍及必要事實；消費者重新核對執行資格，不因持有事件就取得任意資料權限。

**導入條件**

按可重複或亂序投遞設計；明定發布失敗窗口、重試、失敗訊息處置與 schema 演進，不宣稱端到端恰一次。

**驗收**

重複／亂序、舊事件、消費者停機、毒訊息及發布後回應遺失均不重複產生業務效果。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/pubsub/docs/choosing-pubsub-or-cloud-tasks)。[Tasks](background-jobs.md#cloud-tasks) · [公開契約](../../references/ddd/domain-contract.md)。

<a id="eventarc"></a>

## Eventarc

**狀態：需要雲端資源事件觸發時導入。**

**用途與最小流程**

將 Storage 檔案完成上傳等事件導向處理服務。

資源事件 → 明確篩選的 trigger → 授權處理入口 → 核對檔案版本與業務記錄 → 冪等處理。

**資料、授權與失敗邊界**

事件只能證明資源發生變化；須核對 bucket、object、generation 及附件登記，不能接受路徑作為業務授權。

**導入條件**

先選 Standard／Advanced 適用功能、來源及區域；避免處理結果寫回相同資源造成無限事件循環。

**驗收**

重複／亂序事件、未登記檔案、錯誤 bucket、舊 generation 及失敗重試均有處置。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/eventarc/standard/docs/overview)。[Cloud Storage](../storage/cloud-storage.md) · [Functions](#cloud-run-functions)。

<a id="cloud-run-functions"></a>

## Cloud Run functions

**狀態：候選方案；不是現行主服務部署替代品。**

**用途與最小流程**

處理獨立的檔案事件、縮圖、metadata 擷取或輕量整合入口；若既有 Go 服務可簡單處理則沿用。

HTTP 或 CloudEvent 觸發 → 驗證來源及操作身分 → 冪等處理指定資源版本 → 記錄結果。

**資料、授權與失敗邊界**

明確選定執行環境與觸發身分，不把 Firebase Authentication／使用者權限當作自動具備；業務異動仍經相同規則或授權用例。

**導入條件**

依[Cloud Run 決策](../../decisions/0003-cloud-run-primary.md)評估附加 function 的部署與所有權；不是自動拆服務。事件可能重送，限制重試與寫回循環，較長批次改評估 Jobs。

**驗收**

錯誤身分、事件重送、部分失敗、逾時與舊檔案版本不造成重複異動，必要業務操作仍可驗收。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/run/docs/functions/overview)。[Eventarc](#eventarc) · [Jobs](background-jobs.md#cloud-run-jobs)。

<a id="workflows"></a>

## Workflows

**狀態：多步工作需要持久協調時導入。**

**用途與最小流程**

協調 OCR、模型分析、人工確認或結果通知等跨服務步驟。

建立操作 → 依明確輸入呼叫各服務 → 保存步驟結果 → 依錯誤策略重試、等待或結束。

**資料、授權與失敗邊界**

每步用最小服務權限；上下文只放操作 ID 與必要參照，不放大量檔案或長期機密。

**導入條件**

定義期限、重試、取消、回呼驗證及補償；流程完成不能代替領域核准，也不讓工作流直接繞過 Go 規則。

**驗收**

中途失敗、回呼重送、取消後遲到結果、重跑與補償失敗可追查。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/workflows/docs/overview)。[Document AI](document-processing.md#document-ai) · [Tasks](background-jobs.md#cloud-tasks)。

[Google 服務索引與查證範圍](README.md#sources)
