# Cloud Run 部署

依 [ADR 0003](../architecture/decisions/0003-cloud-run-primary.md)，只部署一個 Go HTTP 服務。
根目錄 `cloudbuild.yaml` 使用現有 Dockerfile，依序檢查設定、建置、推送 Artifact Registry、
部署 Cloud Run。機密不進入建置環境；僅由 Cloud Run 引用 Secret Manager。

## 預設配置

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
| 請求 timeout | 15 秒；應用內 LINE 呼叫仍有自己的 5 秒期限 |
| HTTP | 公開 ingress，允許 LINE Webhook 與瀏覽器存取 |
| PORT | 使用 Cloud Run 注入的 8080，不透過自訂 env 覆寫 |

Cloud Run 明確使用 8080 的 startup TCP probe 檢查容器啟動；本專案沒有 /healthz，沒有新增 HTTP health probe。
此配置不包括 Scheduler、Worker、Terraform 或另一套 GitHub 部署 workflow。
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
| 執行服務帳戶 | Secret Manager Secret Accessor，針對部署檔引用的四個 secrets 分別授權 |

建置帳戶不需讀取 LINE 機密內容。執行帳戶不需要部署、建置或 registry 寫入權限。
Cloud Run 透過執行帳戶提供 ADC，不上傳 Google service-account key file。
出勤與群聊摘要已有 Firestore 讀寫實作；執行帳戶需具備目標資料庫所需權限，並以實際讀寫驗收。
`cmd/firestorecheck` 仍使用執行該工具之身分的 ADC，與 Cloud Run 服務是否部署成功分開驗證。

## 設定契約

### Secret Manager 與環境變數

你提供的 Secret Manager 清單如下。`cloudbuild.yaml` 目前只將標為「已引用」的項目注入 Cloud Run；部署檔不寫機密明文：

| Secret Manager 名稱 | 容器環境變數 | 部署檔參照 | 用途／狀態 |
| --- | --- | --- | --- |
| `cwa-api-key` | `CWA_API_KEY` | `1` | 已引用；天氣功能使用。 |
| `gemini-api-key` | `GEMINI_API_KEY` | `latest` | 已引用；群聊摘要使用。 |
| `line-bot-openai-api-key` | 無 | 未引用 | 目前程式與部署檔均未引用。 |
| `line-messaging-api-channel-access-token` | `LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN` | `1` | 已引用；LINE Messaging API 使用。 |
| `line-messaging-api-channel-secret` | `LINE_MESSAGING_API_CHANNEL_SECRET` | `1` | 已引用；LINE Webhook 驗簽使用。 |
| `openai-api-key` | `OPENAI_API_KEY` | `1` | 已引用；OpenAI 群聊摘要使用。 |

「未引用」只表示此儲存庫目前沒有將該 Secret 綁定到 Cloud Run，不代表 Secret 不存在。Secret 內容與版本是否在 Google Cloud 可用，需在雲端另行確認。

`_SUMMARY_PROVIDER=openai` 與 `_SUMMARY_MODEL=gpt-4.1-mini` 透過 `--set-env-vars` 成為容器的 `SUMMARY_PROVIDER`、`SUMMARY_MODEL`，兩者不是機密。摘要還需要 `SUMMARY_TASK_QUEUE`、`SUMMARY_TASK_URL`、`SUMMARY_TASK_SERVICE_ACCOUNT`；這三項的值也由同一個部署檔注入。Cloud Build Trigger 若設定同名 substitutions，會覆寫部署檔中的預設值；請以實際 Build 與 Cloud Run revision 設定驗收。

| Cloud Build substitution | 注入位置／用途 | 預設 |
| --- | --- | --- |
| _MESSAGING_API_CHANNEL_ID | LINE_MESSAGING_API_CHANNEL_ID；必填 | 空，缺少時 preflight 失敗 |
| _MINI_APP_CHANNEL_ID | LINE_MINI_APP_CHANNEL_ID；選用 | 空 |
| _MINI_APP_LIFF_ID | LINE_MINI_APP_LIFF_ID；選用，需與上一項一起提供 | 空 |
| _CHANNEL_SECRET_NAME | LINE_MESSAGING_API_CHANNEL_SECRET 的 secret 名稱 | line-messaging-api-channel-secret |
| _CHANNEL_SECRET_VERSION | 上述 secret 的版本 | 1 |
| _CHANNEL_ACCESS_TOKEN_NAME | LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN 的 secret 名稱 | line-messaging-api-channel-access-token |
| _CHANNEL_ACCESS_TOKEN_VERSION | 上述 secret 的版本 | 1 |

使用版本 1 前須確認已存在且 enabled；若實際版本不同，明確覆寫。
preflight 拒絕 `latest`，避免輪替後同一 revision 的新 instance 取得未記錄的新機密版本。
輪替時新增 secret version，再帶新的數字版本部署；保留舊版本供回滾使用。

GOOGLE_CLOUD_PROJECT 自動設為此次 Build 的 project。MINI App 兩個 ID 全空時停用身分功能；
只提供其中一個會在建置前失敗。Channel、LIFF ID 都不是機密，但必須符合實際環境。
Service Message token、MCP API key、使用者 ID token 與 ADC 憑證不屬於本部署的 substitutions。
應用設定命名仍依 [LINE 契約](../integrations/line-environment.md)。

部署檔使用 set-env-vars／set-secrets 宣告完整集合，會移除未列出的手動 env／secret 綁定；
新增功能時先更新檔案，不能只在 Console 臨時補值。這些值都在執行期注入，沒有 bake 進映像。

## 首次手動部署

先檢查上傳範圍；此命令只列本機檔案：

```bash
bash scripts/tool-exec.sh gcloud meta list-files-for-upload --quiet
```

`.gcloudignore` 只允許 Dockerfile、.dockerignore、cloudbuild.yaml、Go module、cmd/ 與 internal/。
不會上傳 .env、.codex、憑證、工具快取、分析輸出或文件。

確認雲端前置條件與部署授權後，替換下列 project、bucket、Channel ID 及 secret 版本：

```bash
bash scripts/tool-exec.sh gcloud builds submit . \
  --project=YOUR_PROJECT_ID \
  --region=asia-east1 \
  --gcs-source-staging-dir=gs://YOUR_SOURCE_BUCKET/source \
  --config=cloudbuild.yaml \
  --substitutions=_MESSAGING_API_CHANNEL_ID=YOUR_CHANNEL_ID,_CHANNEL_SECRET_VERSION=1,_CHANNEL_ACCESS_TOKEN_VERSION=1 \
  --quiet
```

這個命令會真的建立 Build、上傳映像並部署；不是 dry run。
需要 MINI App 時，在 substitutions 加入 _MINI_APP_CHANNEL_ID 與 _MINI_APP_LIFF_ID。
所有 gcloud 命令以實際安裝版本的 leaf-level help 為準；builds submit／run deploy 沒有一般 dry-run 旗標。

## GitHub 自動部署（手動部署驗收後）

在 Cloud Build Console 連接本 repository、建立 push Trigger，選定一個部署分支，
設定 Build configuration 為根目錄 cloudbuild.yaml，填入相同的非機密 substitutions。
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
