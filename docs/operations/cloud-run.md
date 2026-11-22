# Cloud Run 部署

出勤查詢需要三個複合索引，定義在 [Firestore 索引清單](../../firestore.indexes.json)。自訂 Cloud Build 流程在發布前建立缺少的出勤索引，等待全部達到 `READY`。建置身分需要 Firestore 索引管理權限。使用 inline trigger 或其他部署平台時，在部署步驟前執行同一個檢查：

```sh
python3 scripts/db/ensure-attendance-indexes.py "$PROJECT_ID"
```

指令預設使用 `(default)` 資料庫，可用第二個參數指定資料庫。它從 Cloud Build 執行身分或環境變數 `GOOGLE_OAUTH_ACCESS_TOKEN` 取得授權，只建立出勤索引並檢查狀態。Firebase CLI 也可從根目錄的 `firebase.json` 取得完整索引與欄位設定。

<a id="cloud-run-部署"></a>

```mermaid
flowchart LR
  h["Cloud Run 部署"]
  h --- n0["依 ADR 0003，只部署一個 Go HTTP 服務。"]
  click n0 "../decisions/0003-cloud-run-primary.md"
  h --- n1["Monorepo 候選驗收、裝置清單、無流量發布與具<br/>名 revision 回復見 整合與發布驗收。"]
  click n1 "release-acceptance.md"
```

## 觸發條件

Managed Trigger 的 `includedFiles`／`ignoredFiles` 在 Cloud Build 維護。文件專用變更可依線上規則略過；根目錄 `firestore.indexes.json` 與 `scripts/generate/prepare-miniapp-assets.mts` 已列為觸發輸入。若要讓單改根目錄 `cloudbuild.yaml` 也觸發建置，需另外更新雲端篩選。自訂 Trigger 若指定舊位置，須改指向根目錄的 YAML。

建置不查詢或自動取消其他建置；現行流程不需要額外的查詢或取消權限。映像輸入與本機驗收見[容器建置契約](container-build.md#容器建置契約)。

<a id="建置入口"></a>

```mermaid
flowchart LR
  h["建置入口"]
  h --- n0["現況與提案對照　Managed trigger 與手動提交為<br/>現行入口；自訂 Trigger 為候選方案。"]
```

```mermaid
block-beta
  columns 3
  M["main push：GitHub managed trigger"] H["手動 gcloud builds submit"] C["未來自訂 Trigger"]
  I["inline build 與根 Dockerfile"] Y["cloudbuild.yaml"]:2
  G["image-only update"] D["build → push → deploy"]:2
  M --> I
  H --> Y
  C --> Y
  I --> G
  Y --> D
```

```mermaid
flowchart LR
  h["建置入口"]
  h --- n0["手動入口使用 gcloud builds submit <br/>--config=cloudbuild.yaml。Managed <br/>trigger 的 substitutions、環境變數與服務帳戶<br/>由實際 inline build 和服務設定管理。"]
  h --- n1["現行 managed trigger 的 runtime 環境變數與 <br/>Secret reference 以 Cloud Run service <br/>configuration 為準；image-only update 會沿用<br/>當時的 service template。Git push 不會把 <br/>repository 的 .env.example 或 <br/>cloudbuild.yaml 設定同步到 Cloud Run<br/>。機密只由 Cloud Run 引用 Secret Manager，不<br/>進入 Dockerfile、映像或 Git。"]
```

<a id="managed-trigger-的-rich-menu-圖片路徑"></a>

## Managed trigger 的 Rich Menu 自動同步遷移

原 inline Trigger 使用 /richmenu；新映像已移除此獨立執行檔，改由單一 /server 的部署工作模式執行。**先遷移 Trigger，再部署新映像。** Repository YAML 已包含部署前的單次 Job，不能假設修改 YAML 就會更新線上 inline 設定。接線、租約、權限與尚未完成的雲端驗收見 [Rich Menu 自動發布](rich-menu.md)。

圖片仍位於映像的 /assets/line/richmenu/，目錄可穿越、JPEG 0644，供 nonroot Job 讀取。保留 Dockerfile 的權限整理，不將圖片載入失敗誤判成 LINE 平台問題。

<a id="repository-deploycloudbuildyaml-預設配置"></a>

```mermaid
flowchart LR
  h["Repository #96;cloudbuild.yaml#96; 預設配置"]
  h --- n0["下表描述 repository 配置，不代表現行 managed<br/> trigger 或 serving revision 已套用相同值。"]
```

```mermaid
flowchart LR
  h["Repository #96;cloudbuild.yaml#96; 預設配置"]
  h --- n0["項目：Cloud Run service<br/>預設值：line-mini-app-bot"]
  h --- n1["項目：Region<br/>預設值：asia-east1；仍應依 Firestore 位置確<br/>認"]
  h --- n2["項目：Artifact Registry repository<br/>預設值：applications（Docker 類型，與部署 <br/>region 相同）"]
  h --- n3["項目：映像標記<br/>預設值：Cloud Build 的 BUILD_ID；手動與 <br/>Trigger 都有值，不依賴 COMMIT_SHA"]
  h --- n4["項目：建置服務帳戶 ID<br/>預設值：cloud-build-deployer"]
  h --- n5["項目：應用執行服務帳戶 ID<br/>預設值：line-mini-app-runtime"]
  h --- n6["項目：CPU／記憶體<br/>預設值：1 CPU／512 MiB"]
  h --- n7["項目：Service instance 範圍<br/>預設值：min=0、max=3；不是成本上限或 HA 保證"]
  h --- n8["項目：每 instance concurrency<br/>預設值：20，初始配置，非效能驗收結果"]
  h --- n9["項目：請求 timeout<br/>預設值：100 秒；應用內 LINE 呼叫仍有自己的期<br/>限"]
  h --- n10["項目：HTTP<br/>預設值：公開 ingress，允許 LINE Webhook 與瀏<br/>覽器存取"]
  h --- n11["項目：PORT<br/>預設值：使用 Cloud Run 注入的 8080，不透過自<br/>訂 env 覆寫"]
```

```mermaid
flowchart LR
  h["Repository #96;cloudbuild.yaml#96; 預設配置"]
  h --- n0["Cloud Run 明確使用 8080 的 startup TCP probe<br/> 檢查容器啟動；本專案沒有 /healthz，沒有新增<br/> HTTP health probe。 此配置不包括 Scheduler<br/>、Worker、Terraform 或另一套 GitHub Actions <br/>部署 workflow。 既有 GitHub Repository <br/>checks 保留；Cloud Build Trigger 不會自動等<br/>待 GitHub checks。"]
```

<a id="一次性準備"></a>

```mermaid
flowchart LR
  h["一次性準備"]
  h --- n0["先由操作者選定 project、region 與部署分支，<br/>再在 Google Cloud Console 準備："]
  h --- n1["啟用 Cloud Build、Artifact Registry、Cloud <br/>Run Admin 與 Secret Manager APIs。 真正使用 <br/>Firestore 時再確認 Firestore API、database <br/>與位置；此部署不建立資料庫。"]
  n1 --> n2["建立同 region 的 Docker repository <br/>applications。"]
  n2 --> n3["建立 cloud-build-deployer 與 <br/>line-mini-app-runtime 兩個服務帳戶。 名稱可<br/>透過 substitutions 調整，值為帳戶 ID，不是完<br/>整 email。"]
  n3 --> n4["核對下表六個 Secret references。現有 <br/>repository YAML 固定綁定全部六項， 即使應用<br/>不使用其中一項，執行該 YAML 前仍須能解析每個<br/> reference。 LINE Messaging 需要 Channel <br/>secret 與 access token；摘要依 provider 使用<br/> OpenAI 或 Gemini key， 天氣使用 CWA key，<br/>MINI App Channel secret 目前僅預留。不要為湊<br/>齊部署綁定建立假憑證； 若要省略未使用綁定，<br/>先另行調整部署設定並驗證。記下實際使用的數字<br/>版本。 不把實際值輸入 Cloud Build <br/>substitutions、YAML、Docker build args 或 <br/>Git。"]
  n4 --> n5["若使用手動 source upload，準備專用的來源 GCS<br/> bucket，授予建置帳戶讀取權限。 送出者需具備<br/>建立 Build、使用建置服務帳戶及上傳來源的權限<br/>。"]
```

<a id="兩種服務帳戶的權限"></a>

```mermaid
flowchart LR
  h["兩種服務帳戶的權限"]
  h --- n0["身分：建置服務帳戶<br/>權限與範圍：Artifact Registry Writer，限目標<br/> repository"]
  h --- n1["身分：建置服務帳戶<br/>權限與範圍：Logs Writer，目標 project；使用 <br/>CLOUD_LOGGING_ONLY"]
  h --- n2["身分：建置服務帳戶<br/>權限與範圍：Cloud Run Admin，目標 project，<br/>用於初次建立與設定公開存取；日後可依服務範圍<br/>縮限"]
  h --- n3["身分：建置服務帳戶<br/>權限與範圍：Service Account User，只針對應用<br/>執行服務帳戶"]
  h --- n4["身分：建置服務帳戶<br/>權限與範圍：手動提交時，Storage Object <br/>Viewer，限此次來源 bucket"]
  h --- n5["身分：執行服務帳戶<br/>權限與範圍：Secret Manager Secret Accessor，<br/>針對部署檔引用的六個 secrets 分別授權"]
```

```mermaid
flowchart LR
  h["兩種服務帳戶的權限"]
  h --- n0["建置帳戶不需讀取 LINE 機密內容。執行帳戶不需<br/>要部署、建置或 registry 寫入權限。 Cloud Run<br/> 透過執行帳戶提供 ADC，不上傳 Google <br/>service-account key file。 出勤與群聊摘要已<br/>有 Firestore 讀寫實作；執行帳戶需具備目標資<br/>料庫所需權限，並以實際讀寫驗收。 <br/>check-firestore 子命令 仍使用執行該工具之身分的 <br/>ADC，與 Cloud Run 服務是否部署成功分開驗證。"]
```

<a id="設定契約"></a>
<a id="secret-manager-與環境變數"></a>
<a id="mini-app-公開設定遷移"></a>

## 執行環境設定

Secret Manager 綁定、建置參數與 MINI App 公開設定遷移見[正式環境設定](configuration.md)。

<a id="首次手動部署"></a>

## 首次手動部署

先檢查上傳範圍；此命令只列本機檔案：

```mermaid
flowchart LR
  h["首次手動部署"]
  h --- n0["bash scripts/tooling/tool-exec.sh gcloud meta list-files-for-upload --quiet"]
```

```mermaid
flowchart LR
  h["首次手動部署"]
  h --- n0[".gcloudignore 允許 Dockerfile、.dockerignore<br/>、cloudbuild.yaml、根工作區 manifest<br/>／lock、backend/、frontend/、api/、前端資產<br/>準備腳本與 richmenu 資產；排除依賴目錄及建置<br/>產物。 不會上傳 .env、.codex、憑證、工具快取<br/>、分析輸出或文件。"]
  h --- n1["確認雲端前置條件與部署授權後，替換下列 <br/>project、bucket、Channel ID 及 secret 版本："]
```

```mermaid
flowchart LR
  h["首次手動部署"]
  h --- n0["bash scripts/tooling/tool-exec.sh gcloud builds submit . #92;<br/>  --project=YOUR_PROJECT_ID #92;<br/>  --region=asia-east1 #92;<br/>  --gcs-source-staging-dir=gs://YOUR_SOURCE_BUCKET/source #92;<br/>  --config=cloudbuild.yaml #92;<br/>  --substitutions=_MESSAGING_API_CHANNEL_ID=YOUR_CHANNEL_ID,_CHANNEL_SECRET_VERSION=3,_CHANNEL_ACCESS_TOKEN_VERSION=2 #92;<br/>  --quiet"]
```

```mermaid
flowchart LR
  h["首次手動部署"]
  h --- n0["這個命令會真的建立 Build、上傳映像並部署；不<br/>是 dry run。 MINI App 公開配對已內建；如需切換環境，在 substitutions 設定 _MINI_APP_ENV。 所有 gcloud 命令以實際安<br/>裝版本的 leaf-level help 為準；builds submit<br/>／run deploy 沒有一般 dry-run 旗標。"]
```

<a id="現行-github-managed-trigger"></a>

```mermaid
flowchart LR
  h["現行 GitHub managed trigger"]
  h --- n0["目前 Cloud Build Trigger <br/>rmgpgab-line-mini-app-bot-asia-east1-963sup-line-mini-app-boohl<br/> 只監聽 ^main$，Build configuration 是 <br/>Google 管理的 inline build。Build step 使用 <br/>Dockerfile；Deploy step 只更新 image，因此沿<br/>用 Cloud Run service template 的 runtime 設<br/>定。修改 repository cloudbuild.yaml <br/>或只 push 相同 commit 都不會改變 runtime 設<br/>定； 環境變數或 Secret reference 變更必須在 <br/>Cloud Run 建立新 revision。"]
  h --- n1["截至 2026-11-08，現行服務使用 <br/>185354881750-compute@developer.gserviceaccount.com<br/> 作為 runtime service account；這與 <br/>repository cloudbuild.yaml 預設的 <br/>line-mini-app-runtime 不同。兩者的 IAM 不會<br/>自動同步， 切換部署路徑前必須重新核對 Secret<br/>、Firestore、Cloud Tasks 與 service account <br/>impersonation。"]
  h --- n2["每次 main push 後要同時驗證 Build 的 source <br/>SHA、Deploy step、latest Ready revision、 <br/>traffic 與該 revision 的 env／Secret <br/>reference。Build 綠燈不等於 LINE 驗簽、<br/>Firestore、 Cloud Tasks、模型或訊息送達已通<br/>過端到端測試。"]
```

<a id="docker-buildkit-建置契約"></a>

```mermaid
flowchart LR
  h["Docker BuildKit 建置契約"]
  h --- n0["Dockerfile 使用 RUN --mount=type=secret,<br/>id=build-ca，因此 builder 必須支援並啟用 <br/>BuildKit。 目前 managed trigger 的 inline <br/>Build step 與 repository <br/>cloudbuild.yaml 的 build step 都明確<br/>設定 DOCKER_BUILDKIT=1；Dockerfile 頂端的 <br/>syntax directive 不會自行啟用 BuildKit。 本<br/>機 Docker 預設啟用不代表 Cloud Build 也啟用<br/>。若更新 Docker builder，確認其 BuildKit／<br/>Buildx 支援，並用實際 builder 測試 secret <br/>mount；不能移除 TLS 驗證或將 CA／憑證寫入映<br/>像代替。"]
  h --- n1["若日誌顯示 the --mount option requires <br/>BuildKit，失敗發生在建置階段，尚未推送／部署<br/>新版。 先核對實際 Trigger 的 inline Build <br/>env；只修改 repository YAML 不會更新 managed<br/> trigger。 修改 trigger 前保存 metadata，變<br/>更後比對僅預期的 Build env 有差異，不改部署<br/>、選單或機密設定。"]
  h --- n2["參考：Docker BuildKit、 Docker build secrets<br/>。"]
  n2 --- n2l0["Docker BuildKit"]
  click n2l0 "https://docs.docker.com/build/buildkit/"
  n2 --- n2l1["Docker build secrets"]
  click n2l1 "https://docs.docker.com/build/building/secrets/"
```

<a id="改用-repository-deploycloudbuildyaml-的自訂-trigger"></a>

```mermaid
flowchart LR
  h["改用 repository #96;cloudbuild.yaml#96; 的<br/>自訂 Trigger"]
  h --- n0["在 Cloud Build Console 連接本 repository、建<br/>立 push Trigger，選定一個部署分支， 設定 <br/>Build configuration 為 <br/>cloudbuild.yaml，填入相同的非機密 <br/>substitutions。 建議使用受保護分支，合併前要<br/>求既有 Repository checks 通過；不要設定 PR <br/>自動正式部署。"]
  h --- n1["Trigger 的 Service account 必須明確選擇 <br/>cloud-build-deployer 的完整 email。 Trigger <br/>啟動時會忽略 YAML 的 serviceAccount 欄位，只<br/>採用 Trigger 上設定的帳戶； 只改 <br/>_BUILD_SERVICE_ACCOUNT substitution 不能變更<br/> Trigger 的執行身分。 建置步驟自動推送每個 <br/>BUILD_ID 對應的映像，再部署同一個服務，不建<br/>立 preview app。"]
```

<a id="部署後確認與回滾"></a>

```mermaid
flowchart LR
  h["部署後確認與回滾"]
  h --- n0["依發布驗收與回復流程核對候選、裝置驗收與原流<br/>量分配。Webhook 路徑為 /webhooks/line/{<br/>Messaging API Channel ID}，公開 HTTP 仍須驗<br/>簽。/api/v1/me 驗證身分；瀏覽器使用 Cookie <br/>Session，既有無 Cookie Bearer 路徑驗證 LINE <br/>ID token。"]
  click n0 "release-acceptance.md#切換觀察回復"
  h --- n1["舊版本與映像的保留、清理清單及中止條件見發布<br/>驗收的清理規則。 修改環境與 secret 版本後要<br/>重新驗收。避免同時執行多次部署，以免較舊 <br/>Build 最後完成而覆蓋新版。"]
  click n1 "release-acceptance.md#舊版本與映像清理"
```

<a id="配置驗證狀態與來源"></a>

```mermaid
flowchart LR
  h["配置驗證狀態與來源"]
  h --- n0["本機已以 gcloud 592.0.0 leaf help 核對 <br/>deploy／builds submit 旗標，並以其 Cloud <br/>Build 官方 parser 驗證 YAML 結構；另外檢查 <br/>preflight 與來源上傳清單。 這些檢查不證明 <br/>project、API、IAM、secrets、Trigger 或 Cloud<br/> Run 服務已存在。"]
  h --- n1["Cloud Build substitutions"]
  click n1 "https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values"
  h --- n2["User-specified build service accounts"]
  click n2 "https://docs.cloud.google.com/build/docs/securing-builds/configure-user-specified-service-accounts"
  h --- n3["Deploy to Cloud Run"]
  click n3 "https://docs.cloud.google.com/build/docs/deploying-builds/deploy-cloud-run"
  h --- n4["Cloud Run secrets"]
  click n4 "https://docs.cloud.google.com/run/docs/configuring/services/secrets"
  h --- n5["配置參考 kkdai/linebot-garbage-helper 的 <br/>build → push → deploy 流程； 本專案採 <br/>Artifact Registry、分離服務身分與 Secret <br/>Manager，沒有加入其 token endpoint 或 <br/>Scheduler。"]
```
