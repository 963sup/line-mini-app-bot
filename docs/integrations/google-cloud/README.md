# Google Cloud 與 Maps 整合

依本專案的 LINE Bot、Mini App、工程附件、背景工作及報表需求，一種服務維護一份簡要規格。已存在的 Firestore 與 Cloud Run 文件沿用原位置；Cloud Storage 歸儲存目錄，其餘服務放在本目錄。

## 服務規格

| 服務 | 專案用途 | 採用狀態 |
|---|---|---|
| [Cloud Run](../../operations/cloud-run.md) | 應用 API、LINE Webhook 與容器部署 | 現行部署來源 |
| [Firestore](../firestore/connection.md) | 業務狀態與資料保存 | 現行整合來源 |
| [Cloud Storage](../storage/cloud-storage.md) | 檔案、附件及匯出內容 | 候選方案 |
| [Cloud Tasks](cloud-tasks.md) | 將摘要、通知或文件處理交給指定 HTTP worker，控制重試與速率。 | 已有摘要佇列接線；其他背景工作按需擴充 |
| [Cloud Scheduler](cloud-scheduler.md) | 定時產生日摘要、出勤提醒、檢查或清理工作。 | 按需導入 |
| [Cloud Run Jobs](cloud-run-jobs.md) | 執行批次匯入、匯出、資料回填與統計重建。 | 按需導入 |
| [Secret Manager](secret-manager.md) | 保存 LINE Token、Channel Secret 及服務端 API 機密，支援版本與輪替。 | 部署文件已有規範；雲端狀態另驗收 |
| [IAM 與服務帳戶](iam.md) | 控制建置、部署、應用執行及工作觸發對 Google 資源的存取。 | 部署文件已有規範；雲端狀態另驗收 |
| [Cloud Logging](cloud-logging.md) | 以結構化日誌追查 API、LINE Webhook 與背景工作的成功、拒絕及失敗。 | 營運基礎；採集與保存設定待核驗 |
| [Cloud Monitoring](cloud-monitoring.md) | 觀測可用性、延遲、錯誤率、佇列積壓及處理失敗。 | 營運基礎；指標與告警待核驗 |
| [Cloud Trace](cloud-trace.md) | 定位 HTTP、Firestore、LINE 與 AI 呼叫的延遲及失敗位置。 | 按需導入 |
| [Cloud Build](cloud-build.md) | 執行受控建置、檢查、映像推送及 Cloud Run 發布。 | 已有部署流程來源；目標環境另驗收 |
| [Artifact Registry](artifact-registry.md) | 保存建置映像，讓部署與回復能定位確切產物。 | 已有部署流程來源；目標環境另驗收 |
| [Document AI](document-ai.md) | 從工程 PDF、掃描文件及表格擷取文字與候選欄位。 | 按需導入 |
| [Vertex AI／Google 生成式 AI](vertex-ai.md) | 用於摘要、工程文件理解、分類或檢索輔助，不自行執行業務核准。 | 按需導入；模型與端點於實作時選定 |
| [BigQuery](bigquery.md) | 分析跨專案出勤、工程進度及長期趨勢，減少線上資料的分析掃描。 | 按需導入 |
| [Looker Studio／Data Studio 報表](looker-studio.md) | 呈現經授權的出勤、進度及趨勢管理報表。 | 按需導入 |
| [Pub/Sub](pubsub.md) | 將已發生的業務事實分送通知、統計或搜尋等獨立消費者。 | 多個消費者需要同一事件時導入 |
| [Eventarc](eventarc.md) | 將 Storage 檔案完成上傳等事件導向處理服務。 | 需要雲端資源事件觸發時導入 |
| [Workflows](workflows.md) | 協調 OCR、模型分析、人工確認或結果通知等跨服務步驟。 | 多步工作需要持久協調時導入 |
| [Google Maps Platform](google-maps.md) | 在 Mini App 顯示工地／工區位置、標記及前往地點；地址查找與路線另按需求開啟。 | 按需導入 |
| [Cloud Run functions](cloud-run-functions.md) | 處理獨立的檔案事件、縮圖、metadata 擷取或輕量整合入口；若既有 Go 服務可簡單處理則沿用。 | 候選方案；不是現行主服務部署替代品 |
| [reCAPTCHA](recaptcha.md) | 保護公開表單或高風險操作，提供自動化濫用的風險訊號。 | 公開操作出現濫用風險時導入 |

## 共通責任

Go 持有可信身分、資源授權及業務判定。各服務的接線、身分、資料範圍、失敗策略、成本與驗收於實作時補上具體設定；不因文件存在而建立資源或視為已通過整合驗收。

Maps 是 Google Maps Platform；報表產品及 reCAPTCHA 的名稱與入口依官方當期說明核對。頁面使用專案已討論的名稱，不鎖定未採用的 SDK、模型、價格、區域或方案。新增服務先對照既有 Go 服務的責任，避免重複實作授權或業務規則。

<a id="sources"></a>

## 查證來源與限制

2026-11-22 查閱 Google 官方 Maps、Storage Signed URL、Tasks／Pub/Sub、Scheduler、Run Jobs／functions、Workflows、Eventarc、IAM、Secret Manager、Logging／Monitoring／Trace、Build、Artifact Registry、Document AI、生成式 AI、BigQuery、報表憑證及 reCAPTCHA assessment 文件；各頁連到該服務官方來源。線上文件未固定提交版本，具體 API 與 SDK 於接線時再以專案固定版本核對。

本聊天沒有 Google Developer Knowledge 或 Context7 的原生入口，本次使用官方網站查證。報表舊 overview 入口無法讀取，已改核對官方 Data Studio 資料憑證文件。未查詢雲端實際資源、費率或配額，也未執行部署及功能測試。

[外部整合索引](../README.md) · [架構索引](../../architecture/README.md) · [適用指引](../../architecture/AGENTS.md)
