# Cloud Run 部署

依 [ADR 0003](../architecture/decisions/0003-cloud-run-primary.md)，只部署一個 Go HTTP 服務。

Monorepo 候選驗收、裝置清單、無流量發布與具名 revision 回復見 [整合與發布驗收](release-acceptance.md)。

目前有兩條不同的部署路徑，不能混為一談：

1. 現行 GitHub managed trigger 監聽 `main`，使用 Google 管理的 inline build 與根目錄
   `Dockerfile` 建置映像，再以 image-only update 更新 Cloud Run。它不讀取 repository
   `deploy/cloudbuild.yaml`，也不會把該檔的 substitutions、環境變數或服務帳戶套入服務。
2. `deploy/cloudbuild.yaml` 是手動 `gcloud builds submit --config=deploy/cloudbuild.yaml`
   或未來自訂 Trigger 的完整 build → push → deploy 配置。

現行 managed trigger 的 runtime 環境變數與 Secret reference 以 Cloud Run service
configuration 為準；image-only update 會沿用當時的 service template。Git push 不會把
repository 的 `backend/.env.example` 或 `deploy/cloudbuild.yaml` 設定同步到 Cloud Run。機密只由
Cloud Run 引用 Secret Manager，不進入 Dockerfile、映像或 Git。

### Managed trigger 的 Rich Menu 圖片路徑

現行 managed trigger 的權威設定在 Cloud Build（global trigger
`4276bf03-26ee-4b4b-a206-0d7fc74a9862`）的 inline build；repository
`deploy/cloudbuild.yaml` 沒有 `SyncRichMenus` 步驟，不會覆寫該設定。
inline `SyncRichMenus` 以本次建置映像執行 `/richmenu`，必須傳入
`sync -images /assets/line/richmenu -app-url <MINI App URL> -apply`。
此絕對路徑對應 Dockerfile 的 runtime COPY 目的地；Go CLI 預設的相對路徑
只適合從 repository 根目錄執行，不能用來推定映像內位置。

搬移圖片時須同步檢查 Dockerfile、CLI 預設及實際 trigger 的 `-images` 參數。
2026-11-11 的 `cc82d1e` build 已部署 Ready revision，卻因 trigger 仍傳舊路徑
而在後續 `SyncRichMenus` 報 `attendance-in: provide a readable -image file`。
同日已將實際 trigger 的 `-images` 改為 `/assets/line/richmenu`，讀回核對其餘欄位不變。
Build 失敗不代表前面的部署已回滾；修正 trigger 後仍須以新 build 的同步步驟結果驗收。

## Repository `deploy/cloudbuild.yaml` 預設配置

下表描述 repository 配置，不代表現行 managed trigger 或 serving revision 已套用相同值。

| 項目 | 預設值 |
| --- | --- |
| Cloud Run service | line-mini-app-bot |
| Region | asia-east1；仍應依 Firestore 位置確認 |
| Artifact Registry repository | applications（Docker 類型，與部署 region 相同） |
| 映像標記 | Cloud Build 的 BUILD_ID；手動與 Trigger 都有值，不依賴 COMMIT_SHA |
| 建置服務帳戶 ID | cloud-build-deployer |
| 應用執行服務帳戶 ID | line-mini-app-runtime |
| CPU／記憶體 | 1 CPU／512 MiB |
| Service instance 範圍 | min=0、max=3；不是成本上限或 HA 保證 |
| 每 instance concurrency | 20，初始配置，非效能驗收結果 |
| 請求 timeout | 100 秒；應用內 LINE 呼叫仍有自己的期限 |
| HTTP | 公開 ingress，允許 LINE Webhook 與瀏覽器存取 |
| PORT | 使用 Cloud Run 注入的 8080，不透過自訂 env 覆寫 |

Cloud Run 明確使用 8080 的 startup TCP probe 檢查容器啟動；本專案沒有 /healthz，沒有新增 HTTP health probe。
此配置不包括 Scheduler、Worker、Terraform 或另一套 GitHub Actions 部署 workflow。
既有 GitHub Repository checks 保留；Cloud Build Trigger 不會自動等待 GitHub checks。

## 一次性準備

先由操作者選定 project、region 與部署分支，再在 Google Cloud Console 準備：

1. 啟用 Cloud Build、Artifact Registry、Cloud Run Admin 與 Secret Manager APIs。
   真正使用 Firestore 時再確認 Firestore API、database 與位置；此部署不建立資料庫。
2. 建立同 region 的 Docker repository `applications`。
3. 建立 `cloud-build-deployer` 與 `line-mini-app-runtime` 兩個服務帳戶。
   名稱可透過 substitutions 調整，值為帳戶 ID，不是完整 email。
4. 在 Secret Manager 建立下表兩個 secrets，加入實際 Channel 憑證，記下各自的數字版本。
   不把實際值輸入 Cloud Build substitutions、YAML、Docker build args 或 Git。
5. 若使用手動 source upload，準備專用的來源 GCS bucket，授予建置帳戶讀取權限。
   送出者需具備建立 Build、使用建置服務帳戶及上傳來源的權限。

### 兩種服務帳戶的權限

| 身分 | 權限與範圍 |
| --- | --- |
| 建置服務帳戶 | Artifact Registry Writer，限目標 repository |
| 建置服務帳戶 | Logs Writer，目標 project；使用 CLOUD_LOGGING_ONLY |
| 建置服務帳戶 | Cloud Run Admin，目標 project，用於初次建立與設定公開存取；日後可依服務範圍縮限 |
| 建置服務帳戶 | Service Account User，只針對應用執行服務帳戶 |
| 建置服務帳戶 | 手動提交時，Storage Object Viewer，限此次來源 bucket |
| 執行服務帳戶 | Secret Manager Secret Accessor，針對部署檔引用的六個 secrets 分別授權 |

建置帳戶不需讀取 LINE 機密內容。執行帳戶不需要部署、建置或 registry 寫入權限。
Cloud Run 透過執行帳戶提供 ADC，不上傳 Google service-account key file。
出勤與群聊摘要已有 Firestore 讀寫實作；執行帳戶需具備目標資料庫所需權限，並以實際讀寫驗收。
`cmd/firestorecheck` 仍使用執行該工具之身分的 ADC，與 Cloud Run 服務是否部署成功分開驗證。

## 設定契約

### Secret Manager 與環境變數

Secret Manager 與兩條部署路徑的關係如下。`deploy/cloudbuild.yaml` 的「部署檔參照」只在手動或
自訂 Trigger 真正執行該檔時生效；「現行服務」記錄 2026-11-08 revision
`line-mini-app-bot-00027-695` 的 metadata，不包含 Secret payload：

| Secret Manager 名稱 | 容器環境變數 | 部署檔參照 | 現行服務／用途 |
| --- | --- | --- | --- |
| `cwa-api-key` | `CWA_API_KEY` | `1` | 綁定 `1`；天氣功能使用。 |
| `gemini-api-key` | `GEMINI_API_KEY` | `latest` | 未綁定；Secret 尚無 version，不能解析 `latest`。OpenAI provider 不使用它。 |
| `line-messaging-api-channel-access-token` | `LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN` | `2` | 綁定 `2`；LINE Reply／Push 使用。舊版 `1` 保留。 |
| `line_mini_app_channel_secret` | `LINE_MINI_APP_CHANNEL_SECRET` | `1` | 未綁定；目前 server 未讀取，僅為未來 Service Message 預留。 |
| `line-messaging-api-channel-secret` | `LINE_MESSAGING_API_CHANNEL_SECRET` | `3` | 綁定 `3`；LINE Webhook 驗簽使用。版本 `2` 帶有多餘換行；版本 `1` 已停用。 |
| `openai-api-key` | `OPENAI_API_KEY` | `1` | 綁定 `1`；目前 OpenAI 摘要使用。 |

Secret 名稱存在不代表已有 enabled version。新增或輪替 version 後仍須更新 Cloud Run
reference 並建立 Ready revision；不得用文件或 Secret 名稱推定 payload 正確。

`_SUMMARY_PROVIDER=openai` 與 `_SUMMARY_MODEL=gpt-4.1-mini` 只在執行
`deploy/cloudbuild.yaml` 時透過 `--set-env-vars` 成為 runtime 設定。摘要還需要
`SUMMARY_TASK_QUEUE`、`SUMMARY_TASK_URL`、`SUMMARY_TASK_SERVICE_ACCOUNT`。現行服務已由
Cloud Run service configuration 成套設定 OpenAI provider；queue 層級也已授予 runtime
service account `roles/cloudtasks.enqueuer`，但仍須以真實任務驗收 OIDC、模型與 LINE Push。

| Cloud Build substitution | 注入位置／用途 | 預設 |
| --- | --- | --- |
| _MESSAGING_API_CHANNEL_ID | LINE_MESSAGING_API_CHANNEL_ID；必填 | `2012305727` |
| _MINI_APP_CHANNEL_ID | LINE_MINI_APP_CHANNEL_ID；選用 | `2012412007` |
| _MINI_APP_LIFF_ID | LINE_MINI_APP_LIFF_ID；選用，需與上一項一起提供 | `2012412007-0xCIzCmZ` |
| _CHANNEL_SECRET_NAME | LINE_MESSAGING_API_CHANNEL_SECRET 的 secret 名稱 | line-messaging-api-channel-secret |
| _CHANNEL_SECRET_VERSION | 上述 secret 的版本 | 3 |
| _CHANNEL_ACCESS_TOKEN_NAME | LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN 的 secret 名稱 | line-messaging-api-channel-access-token |
| _CHANNEL_ACCESS_TOKEN_VERSION | 上述 secret 的版本 | 2 |

使用版本 1 前須確認已存在且 enabled；若實際版本不同，明確覆寫。
preflight 拒絕 `latest`，避免輪替後同一 revision 的新 instance 取得未記錄的新機密版本。
輪替時新增 secret version，再帶新的數字版本部署；保留舊版本供回滾使用。

GOOGLE_CLOUD_PROJECT 自動設為此次 Build 的 project。MINI App 兩個 ID 全空時停用身分功能；
只提供其中一個會在建置前失敗。Channel、LIFF ID 都不是機密，但必須符合實際環境。
Service Message token、MCP API key、使用者 ID token 與 ADC 憑證不屬於本部署的 substitutions。
應用設定命名仍依 [LINE 契約](../integrations/line-environment.md)。

`deploy/cloudbuild.yaml` 使用 set-env-vars／set-secrets 宣告完整集合，會移除未列出的手動 env／secret 綁定；
新增功能時先更新檔案，不能只在 Console 臨時補值。這些值都在執行期注入，沒有 bake 進映像。

## 首次手動部署

先檢查上傳範圍；此命令只列本機檔案：

```bash
bash scripts/tooling/tool-exec.sh gcloud meta list-files-for-upload --quiet
```

`.gcloudignore` 允許 Dockerfile、.dockerignore、deploy/cloudbuild.yaml、根工作區 manifest／lock、backend/、frontend/、api/、前端資產準備腳本與 richmenu 資產；排除依賴目錄及建置產物。
不會上傳 .env、.codex、憑證、工具快取、分析輸出或文件。

確認雲端前置條件與部署授權後，替換下列 project、bucket、Channel ID 及 secret 版本：

```bash
bash scripts/tooling/tool-exec.sh gcloud builds submit . \
  --project=YOUR_PROJECT_ID \
  --region=asia-east1 \
  --gcs-source-staging-dir=gs://YOUR_SOURCE_BUCKET/source \
  --config=deploy/cloudbuild.yaml \
  --substitutions=_MESSAGING_API_CHANNEL_ID=YOUR_CHANNEL_ID,_CHANNEL_SECRET_VERSION=3,_CHANNEL_ACCESS_TOKEN_VERSION=2 \
  --quiet
```

這個命令會真的建立 Build、上傳映像並部署；不是 dry run。
需要 MINI App 時，在 substitutions 加入 _MINI_APP_CHANNEL_ID 與 _MINI_APP_LIFF_ID。
所有 gcloud 命令以實際安裝版本的 leaf-level help 為準；builds submit／run deploy 沒有一般 dry-run 旗標。

## 現行 GitHub managed trigger

目前 Cloud Build Trigger `rmgpgab-line-mini-app-bot-asia-east1-963sup-line-mini-app-boohl`
只監聽 `^main$`，Build configuration 是 Google 管理的 inline build。Build step 使用
`Dockerfile`；Deploy step 只更新 image，因此沿用 Cloud Run service template 的 runtime
設定。修改 repository `deploy/cloudbuild.yaml` 或只 push 相同 commit 都不會改變 runtime 設定；
環境變數或 Secret reference 變更必須在 Cloud Run 建立新 revision。

截至 2026-11-08，現行服務使用
`185354881750-compute@developer.gserviceaccount.com` 作為 runtime service account；這與
repository `deploy/cloudbuild.yaml` 預設的 `line-mini-app-runtime` 不同。兩者的 IAM 不會自動同步，
切換部署路徑前必須重新核對 Secret、Firestore、Cloud Tasks 與 service account impersonation。

每次 `main` push 後要同時驗證 Build 的 source SHA、Deploy step、latest Ready revision、
traffic 與該 revision 的 env／Secret reference。Build 綠燈不等於 LINE 驗簽、Firestore、
Cloud Tasks、模型或訊息送達已通過端到端測試。

## Docker BuildKit 建置契約

Dockerfile 使用 `RUN --mount=type=secret,id=build-ca`，因此 builder 必須支援並啟用 BuildKit。
目前 managed trigger 的 inline `Build` step 與 repository `deploy/cloudbuild.yaml` 的 `build` step
都明確設定 `DOCKER_BUILDKIT=1`；Dockerfile 頂端的 syntax directive 不會自行啟用 BuildKit。
本機 Docker 預設啟用不代表 Cloud Build 也啟用。若更新 Docker builder，確認其 BuildKit／Buildx
支援，並用實際 builder 測試 secret mount；不能移除 TLS 驗證或將 CA／憑證寫入映像代替。

若日誌顯示 `the --mount option requires BuildKit`，失敗發生在建置階段，尚未推送／部署新版。
先核對實際 Trigger 的 inline Build env；只修改 repository YAML 不會更新 managed trigger。
修改 trigger 前保存 metadata，變更後比對僅預期的 Build env 有差異，不改部署、選單或機密設定。

參考：[Docker BuildKit](https://docs.docker.com/build/buildkit/)、
[Docker build secrets](https://docs.docker.com/build/building/secrets/)。

## 改用 repository `deploy/cloudbuild.yaml` 的自訂 Trigger

在 Cloud Build Console 連接本 repository、建立 push Trigger，選定一個部署分支，
設定 Build configuration 為 `deploy/cloudbuild.yaml`，填入相同的非機密 substitutions。
建議使用受保護分支，合併前要求既有 Repository checks 通過；不要設定 PR 自動正式部署。

**Trigger 的 Service account 必須明確選擇 cloud-build-deployer 的完整 email。**
Trigger 啟動時會忽略 YAML 的 serviceAccount 欄位，只採用 Trigger 上設定的帳戶；
只改 _BUILD_SERVICE_ACCOUNT substitution 不能變更 Trigger 的執行身分。
建置步驟自動推送每個 BUILD_ID 對應的映像，再部署同一個服務，不建立 preview app。

## 部署後確認與回滾

1. 在 Cloud Run Console 記錄 Ready revision、service URL、映像 digest、執行服務帳戶與 secret 版本。
2. 開啟 `<service-url>/location` 與 `/miniapp`，確認頁面和功能設定狀態。
3. 在 LINE Developers 將 Webhook 設為 `<service-url>/webhooks/line/<Messaging API Channel ID>`，
   驗證後確認一對一／群組公開指令；MINI App Endpoint 設為 `<service-url>/miniapp`。
4. 公開 HTTP 不代表免除應用驗證；Webhook 仍驗簽，/api/v1/me 仍驗 LINE ID token。
5. 出勤與 Firestore 持久化已有程式實作；需實測 LINE 事件、資料庫讀寫與本人紀錄，不能把 revision Ready 當作完成出勤驗收。

回滾時在 Cloud Run Console 的 Revisions 選取前一個已驗收 revision，將流量切回 100%，
再確認 LINE 收訊與頁面。不要刪除其映像或 secret versions，直到回滾保留期結束。
修改環境與 secret 版本後要重新驗收。避免同時執行多次部署，以免較舊 Build 最後完成而覆蓋新版。

## 配置驗證狀態與來源

本機已以 gcloud 592.0.0 leaf help 核對 deploy／builds submit 旗標，並以其 Cloud Build
官方 parser 驗證 YAML 結構；另外檢查 preflight 與來源上傳清單。
這些檢查不證明 project、API、IAM、secrets、Trigger 或 Cloud Run 服務已存在。

- [Cloud Build substitutions](https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values)
- [User-specified build service accounts](https://docs.cloud.google.com/build/docs/securing-builds/configure-user-specified-service-accounts)
- [Deploy to Cloud Run](https://docs.cloud.google.com/build/docs/deploying-builds/deploy-cloud-run)
- [Cloud Run secrets](https://docs.cloud.google.com/run/docs/configuring/services/secrets)

配置參考 kkdai/linebot-garbage-helper 的 build → push → deploy 流程；
本專案採 Artifact Registry、分離服務身分與 Secret Manager，沒有加入其 token endpoint 或 Scheduler。
